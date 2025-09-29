package vm

import (
	"fmt"
	"net/http"

	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/schema"
	"github.com/rancher/steve/pkg/server"
	"github.com/rancher/steve/pkg/stores/proxy"
	"github.com/rancher/wrangler/v3/pkg/schemas"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/config"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/generated/clientset/versioned/scheme"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	vmSchemaID = "kubevirt.io.virtualmachine"
)

var (
	kubevirtSubResouceGroupVersion = k8sschema.GroupVersion{Group: "subresources.kubevirt.io", Version: "v1"}
)

func RegisterSchema(scaled *config.Scaled, server *server.Server, options config.Options) error {

	// import the struct EjectCdRomActionInput to the schema, then the action could use it as input,
	// and because wrangler converts the struct typeName to lower title, so the action input should start with lower case.
	// https://github.com/rancher/wrangler/blob/master/pkg/schemas/reflection.go#L26
	server.BaseSchemas.MustImportAndCustomize(EjectCdRomActionInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(BackupInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(RestoreInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(MigrateInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(CreateTemplateInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(AddVolumeInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(RemoveVolumeInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(CloneInput{}, nil)
	server.BaseSchemas.MustImportAndCustomize(CPUAndMemoryHotplugInput{}, nil)

	kubevirtCache := scaled.VirtFactory.Kubevirt().V1().KubeVirt().Cache()

	vms := scaled.VirtFactory.Kubevirt().V1().VirtualMachine()
	vmis := scaled.VirtFactory.Kubevirt().V1().VirtualMachineInstance()
	vmims := scaled.VirtFactory.Kubevirt().V1().VirtualMachineInstanceMigration()

	vmformatter := &vmformatter{
		vmiCache:  scaled.VirtFactory.Kubevirt().V1().VirtualMachineInstance().Cache(),
		pvcCache:  nil,                    // Will be set if needed
		nodeCache: nil,                    // Will be set if needed
		scCache:   nil,                    // Will be set if needed
		clientSet: kubernetes.Clientset{}, // Will be set if needed
	}

	vmStore := &vmStore{
		Store:    proxy.NewProxyStore(server.ClientFactory, nil, server.AccessSetLookup, nil),
		vms:      scaled.VirtFactory.Kubevirt().V1().VirtualMachine(),
		vmCache:  scaled.VirtFactory.Kubevirt().V1().VirtualMachine().Cache(),
		pvcs:     nil,
		pvcCache: nil,
	}

	// Try to initialize subresource client, but don't fail if it's not available
	var virtSubresourceClient rest.Interface
	copyConfig := rest.CopyConfig(server.RESTConfig)
	copyConfig.GroupVersion = &kubevirtSubResouceGroupVersion
	copyConfig.APIPath = "/apis"
	copyConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	virtSubresourceClient, err := rest.RESTClientFor(copyConfig)
	if err != nil {
		// Log the error but continue without subresource client
		// This allows the app to work even when KubeVirt subresources are not available
		fmt.Printf("Warning: Failed to initialize KubeVirt subresource client: %v\n", err)
		fmt.Println("Continuing without subresource support. VMI operations (pause, unpause, softReboot) will not be available.")
		virtSubresourceClient = nil
	}

	actionHandler := vmActionHandler{
		namespace:                 options.Namespace,
		kubevirtCache:             kubevirtCache,
		vms:                       vms,
		vmis:                      vmis,
		vmCache:                   vms.Cache(),
		vmims:                     vmims,
		vmimsc:                    vmims.Cache(),
		virtSubresourceRestClient: virtSubresourceClient,
	}

	t := schema.Template{
		ID: vmSchemaID,
		Customize: func(apiSchema *types.APISchema) {
			apiSchema.ActionHandlers = map[string]http.Handler{
				startVM:    &actionHandler,
				stopVM:     &actionHandler,
				restartVM:  &actionHandler,
				pauseVM:    &actionHandler,
				unpauseVM:  &actionHandler,
				softReboot: &actionHandler,
			}
			apiSchema.ResourceActions = map[string]schemas.Action{
				startVM:    {},
				stopVM:     {},
				restartVM:  {},
				softReboot: {},
				pauseVM:    {},
				unpauseVM:  {},
			}
		},
		Formatter: vmformatter.formatter,
		Store:     vmStore,
	}

	server.SchemaFactory.AddTemplate(t)

	return nil

}
