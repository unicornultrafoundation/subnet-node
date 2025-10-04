# Rook/Ceph Bootstrap Scripts

This package contains scripts to deploy and operate a local Rook/Ceph cluster for development or production-like testing.

## Files

- `rook.sh`: Main entrypoint to deploy/teardown and check health
- `kubectl_retry.sh`: Small helper to retry `kubectl` commands with backoff

## Prerequisites

- Kubernetes cluster reachable by `kubectl`
- Permissions to create resources in `rook-ceph` namespace
- For loop-backed OSD setup (dev profile):
  - `sudo` privileges
  - `losetup` available on the host

## Profiles

- `prod` (default): Uses manifests under `pkg/k8s/rook/prod`
- `dev`: Uses manifests under `pkg/k8s/rook/dev` and applies a config-override for single-node, minimal quorum, and reduced replication

Select a profile via flag or env:

```bash
# Flag (preferred)
./rook.sh --profile dev deploy

# Short flag
./rook.sh -p prod deploy

# Environment variable
ROOK_PROFILE=dev ./rook.sh deploy
```

## Commands

```text
./rook.sh [--profile prod|dev] <command>

Commands:
  deploy             Deploy Rook/Ceph (CRDs, operator, cluster, pools, FS, StorageClasses)
  teardown           Delete all applied manifests (reverse order)
  health             Check CephCluster and component health
  preflight          Run environment checks prior to deploy
  setup-loop-osd     Create a loop device and patch CephCluster to use it (dev/single-node)
  ceph-status        Run `ceph -s` via the toolbox pod
```

## What `deploy` does

1. Applies core manifests (with fallback to `dev` if a core file is missing):
   - `crds.yaml`, `common.yaml`, `toolbox.yaml`
2. Waits for CRDs to register and operator rollout
3. Deletes immutable StorageClasses if present (to allow re-apply)
4. Applies profile manifests:
   - `operator.yaml`, `cluster.yaml`, `pool.yaml`, `filesystem.yaml`
   - CSI StorageClasses: `csi/rbd/storageclass.yaml`, `csi/cephfs/storageclass.yaml`
5. Runs basic health checks for CephCluster, MDS, and RBD

## Environment Variables

- `ROOK_PROFILE` (default: `prod`): profile to use
- `ROOK_DEPLOY_TIMEOUT` (default: `6000` seconds): health wait loops
- Loop OSD (used by `setup-loop-osd`):
  - `LOOP_OSD_IMAGE_PATH` (default: `/var/lib/rook/rook-osd1.img`)
  - `LOOP_OSD_SIZE` (default: `20G`)
  - `LOOP_OSD_WAIT_TIMEOUT` (default: `300` seconds)

## Examples

```bash
# Deploy minimal single-node dev cluster (creates CephFS and RBD StorageClasses)
./rook.sh --profile dev deploy

# Check cluster health
./rook.sh health

# View ceph status via toolbox
./rook.sh ceph-status

# Create and use loop-backed OSD (useful on Colima/single-node)
./rook.sh --profile dev setup-loop-osd

# Tear everything down
./rook.sh teardown
```

## Behavior & Fallbacks

- Manifest resolution prefers profile/core paths; core files fall back to `pkg/k8s/rook/dev` when missing
- StorageClasses `rook-ceph-block` and `rook-cephfs` are deleted before reapply to bypass immutability
- Health checks are best-effort and will print warnings/timeouts but not hard-fail the script unless a fatal error occurs
- On any unhandled error, the script prints detailed diagnostics from the `rook-ceph` namespace and exits non-zero

## Troubleshooting

- Operator stuck Pending:
  - Check node resources; verify `rook-ceph-operator` logs
- CephCluster not healthy:
  - `./rook.sh health` and `./rook.sh ceph-status` for details
- No OSDs appear (dev/single-node):
  - Use `setup-loop-osd` or configure real devices in `cluster.yaml`
- Discovery not working:
  - Ensure `rook-ceph-discover` DaemonSet has Ready pods

## Notes

- Paths used by this script:
  - Core: `pkg/k8s/rook/core`
  - Dev: `pkg/k8s/rook/dev`
  - Prod: `pkg/k8s/rook/prod`
- The script assumes `kubectl` is configured and points to the target cluster

