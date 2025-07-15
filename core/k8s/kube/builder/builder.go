package builder

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/sirupsen/logrus"
	mani "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	clusterUtil "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/types/v1"
	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	crd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/apis/subnet.node/v1"
)

const (
	SubnetNodeManagedLabelName         = "subnet.node"
	SubnetNodeManifestServiceLabelName = "subnet.node/manifest-service"
	SubnetNodeNetworkStorageClasses    = "subnet.node/storageclasses"
	SubnetNodeServiceTarget            = "subnet.node/service-target"
	SubnetNodeServiceCapabilityGPU     = "subnet.node/capabilities.gpu"
	SubnetNodeServiceCapabilityStorage = "subnet.node/capabilities.storage"
	SubnetNodeMetalLB                  = "metal-lb"
	SubnetNodeDeploymentPolicyName     = "subnet.node-deployment-restrictions"
	SubnetNodeNetworkNamespace         = "subnet.node/namespace"
	SubnetNodeLeaseOwnerLabelName      = "subnet.node/lease.id.owner"
	SubnetNodeLeaseDSeqLabelName       = "subnet.node/lease.id.dseq"
	SubnetNodeLeaseGSeqLabelName       = "subnet.node/lease.id.gseq"
	SubnetNodeLeaseOSeqLabelName       = "subnet.node/lease.id.oseq"
	SubnetNodeLeaseProviderLabelName   = "subnet.node/lease.id.provider"
	SubnetNodeLeaseManifestVersion     = "subnet.node/manifest.version"
	SubnetNodeLeaseUpdatedAt           = "subnet.node/lease.updated_at"
	SubnetNodeManifestResourceVersion  = "subnet.node/manifest.resource.version"
)

const (
	ValTrue  = "true"
	ValFalse = "false"
)

const (
	runtimeClassNoneValue = "none"
	runtimeClassNvidia    = "nvidia"
)

const (
	envVarSubnetNodeGroupSequence         = "SUBNET_NODE_GROUP_SEQUENCE"
	envVarSubnetNodeDeploymentSequence    = "SUBNET_NODE_DEPLOYMENT_SEQUENCE"
	envVarSubnetNodeOrderSequence         = "SUBNET_NODE_ORDER_SEQUENCE"
	envVarSubnetNodeOwner                 = "SUBNET_NODE_OWNER"
	envVarSubnetNodeProvider              = "SUBNET_NODE_PROVIDER"
	envVarSubnetNodeClusterPublicHostname = "SUBNET_NODE_CLUSTER_PUBLIC_HOSTNAME"
)

var (
	ErrKubeBuilder              = errors.New("kube-builder")
	ErrManifestServiceMismatch  = errors.New("manifest service mismatch")
	ErrManifestRenameNotAllowed = errors.New("manifest service rename is not allowed")
)

var (
	dnsPort     = intstr.FromInt(53)
	udpProtocol = corev1.Protocol("UDP")
	tcpProtocol = corev1.Protocol("TCP")
)

type IClusterDeployment interface {
	LeaseID() mtypes.LeaseID
	ManifestGroup() *mani.Group
	UpdateManifest() bool
	ClusterParams() crd.ClusterSettings
	SetResourceVersion(string)
	GetResourceVersion() string
}

type ClusterDeployment struct {
	Lid             mtypes.LeaseID
	Group           *mani.Group
	Sparams         crd.ClusterSettings
	resourceVersion string
	updateManifest  bool
}

var _ IClusterDeployment = (*ClusterDeployment)(nil)

func ClusterDeploymentFromDeployment(d ctypes.IDeployment) (IClusterDeployment, error) {
	var cparams crd.ClusterSettings

	updateManifest := false

	switch sparams := d.ClusterParams().(type) {
	case crd.ClusterSettings:
		cparams = sparams
	case crd.ReservationClusterSettings:
		var err error
		cparams, err = buildClusterSettings(d.ManifestGroup(), sparams)
		if err != nil {
			return nil, err
		}

		updateManifest = true
	default:
		return nil, fmt.Errorf("%w: ClusterParams() returned result of unexpected type (%s)",
			ErrKubeBuilder,
			reflect.TypeOf(sparams),
		)
	}

	cd := &ClusterDeployment{
		Lid:             d.LeaseID(),
		Group:           d.ManifestGroup(),
		Sparams:         cparams,
		resourceVersion: d.ResourceVersion(),
		updateManifest:  updateManifest,
	}

	if err := cd.validate(); err != nil {
		return cd, err
	}

	return cd, nil
}

func buildClusterSettings(mgroup *mani.Group, reservation crd.ReservationClusterSettings) (crd.ClusterSettings, error) {
	sparams := make([]*crd.SchedulerParams, 0, len(mgroup.Services))

	for _, svc := range mgroup.Services {
		params, exists := reservation[svc.Resources.ID]
		if !exists {
			return crd.ClusterSettings{}, fmt.Errorf("%w: reservation does not have SchedulerParams for ResourcesID (%d)", ErrKubeBuilder, svc.Resources.ID)
		}

		sparams = append(sparams, params)
	}

	res := crd.ClusterSettings{
		SchedulerParams: sparams,
	}

	return res, nil
}

func (d *ClusterDeployment) LeaseID() mtypes.LeaseID {
	return d.Lid
}

func (d *ClusterDeployment) ManifestGroup() *mani.Group {
	return d.Group
}

func (d *ClusterDeployment) ClusterParams() crd.ClusterSettings {
	return d.Sparams
}

func (d *ClusterDeployment) SetResourceVersion(val string) {
	d.resourceVersion = val
}

func (d *ClusterDeployment) GetResourceVersion() string {
	return d.resourceVersion
}

func (d *ClusterDeployment) UpdateManifest() bool {
	return d.updateManifest
}

func (d *ClusterDeployment) validate() error {
	if len(d.Group.Services) != len(d.Sparams.SchedulerParams) {
		return fmt.Errorf("%w: group services count does not match scheduler params count (%d) != (%d)",
			ErrKubeBuilder,
			len(d.Group.Services),
			len(d.Sparams.SchedulerParams))
	}

	for idx := range d.Group.Services {
		svc := d.Group.Services[idx]
		sParams := d.Sparams.SchedulerParams[idx]

		if svc.Resources.CPU == nil {
			return fmt.Errorf("%w: service %s. resource CPU cannot be nil", ErrKubeBuilder, svc.Name)
		}

		if svc.Resources.GPU == nil {
			return fmt.Errorf("%w: service %s. resource GPU cannot be nil", ErrKubeBuilder, svc.Name)
		}

		if svc.Resources.Memory == nil {
			return fmt.Errorf("%w: service %s. resource Memory cannot be nil", ErrKubeBuilder, svc.Name)
		}

		if svc.Resources.GPU.Units.Value() > 0 {
			if sParams == nil ||
				sParams.Resources == nil ||
				sParams.Resources.GPU == nil {
				return fmt.Errorf("%w: service %s. SchedulerParams.Resources.GPU must not be nil when GPU > 0", ErrKubeBuilder, svc.Name)
			}
		}
	}

	return nil
}

type builderBase interface {
	NS() string
	Name() string
	Validate() error
	IsObjectRevisionLatest(map[string]string) bool
}

type builder struct {
	log        *logrus.Logger
	settings   Settings
	deployment IClusterDeployment
	mani       *crd.Manifest
	group      mani.Group
	sparams    []*crd.SchedulerParams
}

var _ builderBase = (*builder)(nil)

func (b *builder) NS() string {
	return LidNS(b.deployment.LeaseID())
}

func (b *builder) Name() string {
	return b.NS()
}

func (b *builder) labels() map[string]string {
	res := map[string]string{
		SubnetNodeManagedLabelName: "true",
		SubnetNodeNetworkNamespace: LidNS(b.deployment.LeaseID()),
	}

	res[SubnetNodeManifestResourceVersion] = b.deployment.GetResourceVersion()

	return res
}

func (b *builder) selectorLabels() map[string]string {
	res := map[string]string{
		SubnetNodeManagedLabelName: "true",
		SubnetNodeNetworkNamespace: LidNS(b.deployment.LeaseID()),
	}

	return res
}

func (b *builder) Validate() error {
	return nil
}

func (b *builder) IsObjectRevisionLatest(labels map[string]string) bool {
	objVersion, exists := labels[SubnetNodeManifestResourceVersion]
	if !exists {
		return false
	}

	return b.deployment.GetResourceVersion() == objVersion
}

func addIfNotPresent(envVarsAlreadyAdded map[string]int, env []corev1.EnvVar, key string, value interface{}) []corev1.EnvVar {
	_, exists := envVarsAlreadyAdded[key]
	if exists {
		return env
	}

	env = append(env, corev1.EnvVar{Name: key, Value: fmt.Sprintf("%v", value)})
	return env
}

const SuffixForNodePortServiceName = "-np"

func makeGlobalServiceNameFromBasename(basename string) string {
	return fmt.Sprintf("%s%s", basename, SuffixForNodePortServiceName)
}

// LidNS generates a unique sha256 sum for identifying a provider's object name.
func LidNS(lid mtypes.LeaseID) string {
	return clusterUtil.LeaseIDToNamespace(lid)
}

func AppendLeaseLabels(lid mtypes.LeaseID, labels map[string]string) map[string]string {
	labels[SubnetNodeLeaseOwnerLabelName] = lid.Owner
	labels[SubnetNodeLeaseDSeqLabelName] = strconv.FormatUint(lid.DSeq, 10)
	labels[SubnetNodeLeaseGSeqLabelName] = strconv.FormatUint(uint64(lid.GSeq), 10)
	labels[SubnetNodeLeaseOSeqLabelName] = strconv.FormatUint(uint64(lid.OSeq), 10)
	labels[SubnetNodeLeaseProviderLabelName] = lid.Provider

	return labels
}

func updateSubnetLabels(currLabels, newLabels map[string]string) map[string]string {
	res := make(map[string]string)

	for name, val := range currLabels {
		if !strings.HasPrefix(name, SubnetNodeManagedLabelName) {
			res[name] = val
		}
	}

	for name, val := range newLabels {
		res[name] = val
	}

	return res
}
