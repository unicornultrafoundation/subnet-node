#!/bin/bash -E

CURDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
source "$CURDIR"/kubectl_retry.sh

ROOK_DEPLOY_TIMEOUT=${ROOK_DEPLOY_TIMEOUT:-6000}
rootdir="$(dirname "$0")/../.."

# Default profile if none provided via flags or env
ROOK_PROFILE_DEFAULT=prod

# Parse flags anywhere in the argv and rebuild positional args
ARGS=()
while [[ $# -gt 0 ]]; do
	case "$1" in
		--profile=*)
			ROOK_PROFILE="${1#*=}"
			shift
			;;
		--profile)
			ROOK_PROFILE="$2"
			shift 2
			;;
		-p)
			ROOK_PROFILE="$2"
			shift 2
			;;
		*)
			ARGS+=("$1")
			shift
			;;
	esac
done
# restore remaining args as positionals for the command switch
set -- "${ARGS[@]}"
# Ensure default profile when none provided via flags or env
if [ -z "$ROOK_PROFILE" ]; then
  ROOK_PROFILE="$ROOK_PROFILE_DEFAULT"
fi

# Core manifests now live under pkg/k8s/rook/core; fallback to dev for missing files
ROOK_CORE_PATH=${rootdir}/pkg/k8s/rook/core
ROOK_DEV_PATH=${rootdir}/pkg/k8s/rook/dev
ROOK_PROFILE_PATH=${rootdir}/pkg/k8s/rook/${ROOK_PROFILE}

resolve_manifest_path() {
	# $1 is an absolute path under ROOK_CORE_PATH or ROOK_PROFILE_PATH
	local path="$1"
	if [ -f "$path" ]; then
		echo "$path"
		return 0
	fi
	# fallback core->dev for core files
	if [[ "$path" == "$ROOK_CORE_PATH"/* ]]; then
		local fallback="${path/$ROOK_CORE_PATH/$ROOK_DEV_PATH}"
		if [ -f "$fallback" ]; then
			echo "$fallback"
			return 0
		fi
	fi
	# not found
	return 1
}

if [ -z "$ROOK_CORE_PATH" ] || [ -z "$ROOK_PROFILE_PATH" ]; then
	echo "ROOK paths are not set"
	exit 1
fi

core_files=(
	"${ROOK_CORE_PATH}/crds.yaml"
	"${ROOK_CORE_PATH}/common.yaml"
	"${ROOK_CORE_PATH}/toolbox.yaml"
)

# Profile-specific manifests (cluster/pool/filesystem/storageclasses)
profile_files=(
	"${ROOK_PROFILE_PATH}/operator.yaml"
	"${ROOK_PROFILE_PATH}/cluster.yaml"
	"${ROOK_PROFILE_PATH}/pool.yaml"
	"${ROOK_PROFILE_PATH}/filesystem.yaml"
	"${ROOK_PROFILE_PATH}/csi/rbd/storageclass.yaml"
	"${ROOK_PROFILE_PATH}/csi/cephfs/storageclass.yaml"
)

# For dev profile, also apply config override to set pool defaults to 1
if [[ "$ROOK_PROFILE" == "dev" ]]; then
	core_files+=("${ROOK_DEV_PATH}/config-override.yaml")
fi

trap log_errors ERR

# log_errors is called on exit (see 'trap' above) and tries to provide
# sufficient information to debug deployment problems
function log_errors() {
	# enable verbose execution
	set -x

	kubectl get nodes
	kubectl -n rook-ceph get events
	kubectl -n rook-ceph describe pods
	kubectl -n rook-ceph logs -l app=rook-ceph-operator
	kubectl -n rook-ceph get CephClusters -oyaml
	kubectl -n rook-ceph get CephFilesystems -oyaml
	kubectl -n rook-ceph get CephBlockPools -oyaml

	# this function should not return, a fatal error was caught!
	exit 1
}

function deploy_rook() {
	# Apply core manifests first (with fallback)
	for f in "${core_files[@]}"; do
		if resolved=$(resolve_manifest_path "$f"); then
			kubectl_retry apply -f "$resolved"
		else
			echo "Missing manifest: $f" >&2
			return 1
		fi
	done
	# Wait for CRDs to register before applying CRs
	local crd_wait=0; local crd_timeout=60
	while (( crd_wait < crd_timeout )); do
		if kubectl get crd cephclusters.ceph.rook.io >/dev/null 2>&1; then
			break
		fi
		sleep 2; crd_wait=$((crd_wait+2))
	done
	# Wait for operator to be ready
	kubectl -n rook-ceph rollout status deploy/rook-ceph-operator --timeout=300s || true

	# Delete immutable StorageClasses before re-applying profile SCs and apply profile manifests (with fallback)
	for f in "${profile_files[@]}"; do
		case "$f" in
			*"csi/rbd/storageclass.yaml")
				kubectl get storageclass rook-ceph-block &>/dev/null && kubectl delete storageclass rook-ceph-block || true
				;;
			*"csi/cephfs/storageclass.yaml")
				kubectl get storageclass rook-cephfs &>/dev/null && kubectl delete storageclass rook-cephfs || true
				;;
		esac
		if resolved=$(resolve_manifest_path "$f"); then
			kubectl_retry apply -f "$resolved"
		else
			echo "Missing manifest: $f" >&2
			return 1
		fi
	done

	# Health checks
	if ! kubectl_retry -n rook-ceph get cephclusters -oyaml | grep 'items: \[\]' &>/dev/null; then
		if check_ceph_cluster_health; then
			echo "CEPH cluster is healthy"
		else
			echo "CEPH cluster not in a healthy state (timeout)"
		fi
	fi
	if ! kubectl_retry -n rook-ceph get cephfilesystems -oyaml | grep 'items: \[\]' &>/dev/null; then
		check_mds_stat
	fi
	if ! kubectl_retry -n rook-ceph get cephblockpools -oyaml | grep 'items: \[\]' &>/dev/null; then
		check_rbd_stat ""
	fi
}

function teardown_rook() {
	echo "Tearing down Rook/Ceph Kubernetes resources..."
	
	# Delete profile files first (reverse), then core (with fallback resolution)
	for ((idx=${#profile_files[@]}-1 ; idx>=0 ; idx--)) ; do
		f="${profile_files[idx]}"
		if resolved=$(resolve_manifest_path "$f"); then
			kubectl_retry delete -f "$resolved" || true
		fi
	done
	for ((idx=${#core_files[@]}-1 ; idx>=0 ; idx--)) ; do
		f="${core_files[idx]}"
		if resolved=$(resolve_manifest_path "$f"); then
			kubectl_retry delete -f "$resolved" || true
		fi
	done
	
	echo "Cleaning up physical storage..."
	cleanup_physical_storage
}

function cleanup_physical_storage() {
	# Clean up loop devices created by setup-loop-osd
	if [ -n "${LOOP_OSD_IMAGE_PATH:-}" ] && [ -f "${LOOP_OSD_IMAGE_PATH}" ]; then
		echo "Removing loop device for: ${LOOP_OSD_IMAGE_PATH}"
		
		# Find and remove loop devices associated with this image
		for loop_dev in /dev/loop*; do
			if [ -b "$loop_dev" ]; then
				# Check if this loop device is using our image
				if losetup "$loop_dev" 2>/dev/null | grep -q "${LOOP_OSD_IMAGE_PATH}"; then
					echo "Removing loop device: $loop_dev"
					# Unmount if mounted
					mount | grep "$loop_dev" | awk '{print $3}' | xargs -r sudo umount 2>/dev/null || true
					# Remove loop device
					sudo losetup -d "$loop_dev" 2>/dev/null || true
				fi
			fi
		done
		
		# Remove the image file
		echo "Removing loop OSD image: ${LOOP_OSD_IMAGE_PATH}"
		sudo rm -f "${LOOP_OSD_IMAGE_PATH}" 2>/dev/null || true
	fi
	
	# Clean up any remaining loop devices that might be orphaned
	echo "Checking for orphaned loop devices..."
	for loop_dev in /dev/loop*; do
		if [ -b "$loop_dev" ]; then
			# Check if it's still in use
			if ! losetup "$loop_dev" >/dev/null 2>&1; then
				continue
			fi
			
			# Check if it's associated with Ceph/Rook
			loop_info=$(losetup "$loop_dev" 2>/dev/null)
			if echo "$loop_info" | grep -q "ceph\|rook\|osd"; then
				echo "Removing orphaned Ceph loop device: $loop_dev"
				mount | grep "$loop_dev" | awk '{print $3}' | xargs -r sudo umount 2>/dev/null || true
				sudo losetup -d "$loop_dev" 2>/dev/null || true
			fi
		fi
	done
	
	# Clean up any Ceph-related mount points
	echo "Cleaning up Ceph mount points..."
	mount | grep -E "(ceph|rook|osd)" | awk '{print $3}' | while read mount_point; do
		if [ -d "$mount_point" ]; then
			echo "Unmounting: $mount_point"
			sudo umount "$mount_point" 2>/dev/null || true
		fi
	done
	
	echo "Physical storage cleanup completed"
}

function check_ceph_cluster_health() {
	for ((retry = 0; retry <= ROOK_DEPLOY_TIMEOUT; retry = retry + 5)); do
		CEPH_STATE=$(kubectl_retry -n rook-ceph get cephclusters -o jsonpath='{.items[0].status.state}')
		CEPH_HEALTH=$(kubectl_retry -n rook-ceph get cephclusters -o jsonpath='{.items[0].status.ceph.health}')
		echo "Checking CEPH cluster state: [$CEPH_STATE]"
		if [ "$CEPH_STATE" = "Created" ]; then
			if [ "$CEPH_HEALTH" = "HEALTH_OK" ]; then
				echo "The CEPH cluster health state [$CEPH_HEALTH]"
				break
			elif [ "$retry" -lt "$ROOK_DEPLOY_TIMEOUT" ]; then
				sleep 5
			fi
		fi
	done

	if [ "$retry" -gt "$ROOK_DEPLOY_TIMEOUT" ]; then
		return 1
	fi

	return 0
}

function check_mds_stat() {
	for ((retry = 0; retry <= ROOK_DEPLOY_TIMEOUT; retry = retry + 5)); do
		FS_NAME=$(kubectl_retry -n rook-ceph get cephfilesystems.ceph.rook.io -ojsonpath='{.items[0].metadata.name}')
		echo "Checking MDS ($FS_NAME) stats... ${retry}s" && sleep 5

		ACTIVE_COUNT=$(kubectl_retry -n rook-ceph get cephfilesystems "$FS_NAME" -ojsonpath='{.spec.metadataServer.activeCount}')

		ACTIVE_COUNT_NUM=$((ACTIVE_COUNT + 0))
		echo "MDS ($FS_NAME) active_count: [$ACTIVE_COUNT_NUM]"
		if ((ACTIVE_COUNT_NUM < 1)); then
			continue
		else
			if kubectl_retry -n rook-ceph get pod -l rook_file_system="$FS_NAME" | grep Running &>/dev/null; then
				echo "Filesystem ($FS_NAME) is successfully created..."
				break
			fi
		fi
	done

	if [ "$retry" -gt "$ROOK_DEPLOY_TIMEOUT" ]; then
		echo "[Timeout] Failed to get ceph filesystem pods"
		return 1
	fi
	echo ""
}

function check_rbd_stat() {
	for ((retry = 0; retry <= ROOK_DEPLOY_TIMEOUT; retry = retry + 5)); do
		if [ -z "$1" ]; then
			RBD_POOL_NAME=$(kubectl_retry -n rook-ceph get cephblockpools -ojsonpath='{.items[0].metadata.name}')
		else
			RBD_POOL_NAME=$1
		fi
		echo "Checking RBD ($RBD_POOL_NAME) stats... ${retry}s" && sleep 5

		TOOLBOX_POD=$(kubectl_retry -n rook-ceph get pods -l app=rook-ceph-tools -o jsonpath='{.items[0].metadata.name}')
		TOOLBOX_POD_STATUS=$(kubectl_retry -n rook-ceph get pod "$TOOLBOX_POD" -ojsonpath='{.status.phase}')
		[[ "$TOOLBOX_POD_STATUS" != "Running" ]] && \
			{ echo "Toolbox POD ($TOOLBOX_POD) status: [$TOOLBOX_POD_STATUS]"; continue; }

		if kubectl_retry exec -n rook-ceph "$TOOLBOX_POD" -it -- rbd pool stats "$RBD_POOL_NAME" &>/dev/null; then
			echo "RBD ($RBD_POOL_NAME) is successfully created..."
			break
		fi
	done

	if [ "$retry" -gt "$ROOK_DEPLOY_TIMEOUT" ]; then
		echo "[Timeout] Failed to get RBD pool stats"
		return 1
	fi
	echo ""
}

function setup_loop_osd() {
	# Create a loop-backed device and configure CephCluster to use it
	# Configurable via env:
	#   LOOP_OSD_IMAGE_PATH (default: /var/lib/rook/rook-osd1.img)
	#   LOOP_OSD_SIZE       (default: 20G)
	#   LOOP_OSD_DEVICE     (default: detected via losetup)

	local image_path=${LOOP_OSD_IMAGE_PATH:-/var/lib/rook/rook-osd1.img}
	local size=${LOOP_OSD_SIZE:-20G}

	# Ensure operator config (loop devices + discovery) is applied and operator restarted
	if resolved=$(resolve_manifest_path "${ROOK_CORE_PATH}/operator.yaml"); then
		kubectl_retry apply -f "$resolved"
	fi
	kubectl -n rook-ceph rollout restart deploy/rook-ceph-operator || true
	kubectl -n rook-ceph rollout status deploy/rook-ceph-operator --timeout=120s || true

	# Ensure toolbox and other manifests exist before OSD provisioning
	if resolved=$(resolve_manifest_path "${ROOK_CORE_PATH}/toolbox.yaml"); then
		kubectl_retry apply -f "$resolved"
	fi

	# Create loopback file and attach to a loop device
	sudo mkdir -p /var/lib/rook || true
	if [ ! -f "$image_path" ]; then
		echo "Creating loopback image at $image_path of size $size"
		sudo fallocate -l "$size" "$image_path"
	else
		echo "Loopback image already exists at $image_path"
	fi

	# Find existing loop device for this image or create a new one
	local loop_dev
	loop_dev=$(sudo losetup -j "$image_path" | awk -F: 'NR==1 {print $1}')
	if [ -z "$loop_dev" ]; then
		loop_dev=$(sudo losetup -f --show "$image_path")
	fi
	# Ensure partitions are re-read (even if none present)
	sudo losetup -P "$loop_dev" || true
	echo "Using loop device: $loop_dev"

	# Determine node name
	local node_name
	node_name=$(kubectl_retry get nodes -o jsonpath='{.items[0].metadata.name}')
	[ -z "$node_name" ] && { echo "Failed to detect node name"; return 1; }

	# Patch CephCluster to use the loop device explicitly
	local patch
	patch=$(cat <<PATCH
{"spec":{"storage":{"useAllNodes":false,"useAllDevices":false,"nodes":[{"name":"$node_name","devices":[{"name":"$loop_dev"}]}]}}}
PATCH
)

	echo "Patching CephCluster to use $loop_dev on node $node_name"
	kubectl_retry -n rook-ceph patch cephcluster rook-ceph --type merge -p "$patch"

	# Wait for an OSD to come up
	local waited=0
	local timeout=${LOOP_OSD_WAIT_TIMEOUT:-300}
	while (( waited < timeout )); do
		if kubectl -n rook-ceph get pods -l app=rook-ceph-osd 2>/dev/null | grep -q "rook-ceph-osd"; then
			# Check if at least one OSD pod is Ready 2/2
			if kubectl -n rook-ceph get pods -l app=rook-ceph-osd | awk 'NR>1 {print $2}' | grep -q "^2/2$"; then
				echo "OSD is running and Ready"
				break
			fi
		fi
		sleep 5
		waited=$(( waited + 5 ))
	done

	if (( waited >= timeout )); then
		echo "[Timeout] OSD did not become Ready within $timeout seconds"
		return 1
	fi

	# Show ceph status if toolbox is available
	local toolbox
	toolbox=$(kubectl_retry -n rook-ceph get pods -l app=rook-ceph-tools -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
	if [ -n "$toolbox" ]; then
		kubectl_retry -n rook-ceph exec "$toolbox" -- ceph -s || true
	fi
}

function ceph_status() {
	local toolbox
	toolbox=$(kubectl_retry -n rook-ceph get pods -l app=rook-ceph-tools -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
	if [ -z "$toolbox" ]; then
		echo "rook-ceph-tools pod not found"
		return 1
	fi
	kubectl_retry -n rook-ceph exec "$toolbox" -- ceph -s
}

# Preflight: check environment prerequisites for deploying Rook/Ceph
function preflight() {
	local ok=true
	local missing=()
	for cmd in kubectl; do
		if ! command -v "$cmd" >/dev/null 2>&1; then
			missing+=("$cmd")
		fi
	done
	if (( ${#missing[@]} )); then
		echo "Missing required commands: ${missing[*]}" >&2
		ok=false
	fi
	if ! kubectl cluster-info >/dev/null 2>&1; then
		echo "kubectl cannot reach the cluster (kubectl cluster-info failed)" >&2
		ok=false
	fi
	local ready_nodes
	ready_nodes=$(kubectl get nodes -o jsonpath='{range .items[*]}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}' 2>/dev/null | grep -c True || true)
	if [[ -z "$ready_nodes" || "$ready_nodes" -lt 1 ]]; then
		echo "No Ready Kubernetes nodes found" >&2
		ok=false
	fi
	if ! kubectl auth can-i create namespaces >/dev/null 2>&1; then
		echo "Warning: current user may not create namespaces" >&2
	fi
	if ! kubectl auth can-i create -n rook-ceph deployment >/dev/null 2>&1; then
		echo "Warning: may not create resources in namespace rook-ceph" >&2
	fi
	echo "Using profile: $ROOK_PROFILE"
	for f in "${core_files[@]}"; do
		if ! resolved=$(resolve_manifest_path "$f"); then
			echo "Missing manifest: $f (and no fallback)" >&2
			ok=false
		fi
	done
	for f in "${profile_files[@]}"; do
		if ! resolved=$(resolve_manifest_path "$f"); then
			echo "Missing manifest: $f" >&2
			ok=false
		fi
	done
	# Use profile-specific cluster.yaml
	local cluster_yaml="${ROOK_PROFILE_PATH}/cluster.yaml"
	if resolved=$(resolve_manifest_path "$cluster_yaml"); then
		local use_all_devices has_nodes_devices
		use_all_devices=$(grep -E "^[[:space:]]*useAllDevices:[[:space:]]*true" "$resolved" || true)
		has_nodes_devices=$(awk '/^[[:space:]]*storage:/,0' "$resolved" | grep -E "^[[:space:]]*devices:[[:space:]]*" || true)
		if [[ -z "$use_all_devices" && -z "$has_nodes_devices" ]]; then
			echo "Warning: cluster.yaml does not enable useAllDevices and has no nodes.devices configured. OSDs may not provision." >&2
		fi
	else
		echo "Warning: cluster.yaml not found for profile at $cluster_yaml" >&2
	fi
	if kubectl -n rook-ceph get ds rook-ceph-discover >/dev/null 2>&1; then
		local ds_ready
		ds_ready=$(kubectl -n rook-ceph get ds rook-ceph-discover -o jsonpath='{.status.numberReady}' 2>/dev/null || echo 0)
		if [[ "$ds_ready" == "0" || -z "$ds_ready" ]]; then
			echo "Warning: rook-ceph-discover DaemonSet has no ready pods; device discovery may not work" >&2
		fi
	fi
	if [[ "$ROOK_PROFILE" == "dev" ]]; then
		if ! command -v losetup >/dev/null 2>&1; then
			echo "Warning: losetup not found; setup-loop-osd will fail" >&2
		fi
		if ! sudo -n true >/dev/null 2>&1; then
			echo "Warning: sudo passwordless not available; setup-loop-osd may prompt for password" >&2
		fi
	fi
	if [[ "$ok" == true ]]; then
		echo "Preflight checks PASSED"
		return 0
	else
		echo "Preflight checks FAILED"
		return 1
	fi
}

case "${1:-}" in
deploy)
	deploy_rook
	;;
teardown)
	teardown_rook
	;;
health)
	check_ceph_cluster_health
	;;
preflight)
	preflight
	;;
setup-loop-osd)
	setup_loop_osd
	;;
ceph-status)
	ceph_status
	;;
*)
	echo " $0 [--profile prod|dev] [command]
Available Commands:
  deploy             Deploy a rook
  teardown           Teardown a rook
  health             Check cluster health
  preflight          Run environment checks for deploying Rook/Ceph
  setup-loop-osd     Create a loop device and configure cluster to use it (for Colima/single-node)
  ceph-status        Show 'ceph -s' from toolbox" >&2
	;;
esac
