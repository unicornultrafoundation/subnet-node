package vm

import (
	"net/http"

	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/schema"
	"github.com/rancher/steve/pkg/server"
	"github.com/rancher/steve/pkg/stores/proxy"
	"github.com/rancher/wrangler/v3/pkg/schemas"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/config"
)

const (
	vmSchemaID = "kubevirt.io.virtualmachine"
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
	// TODO: add other resources later

	vmStore := &vmStore{
		Store:    proxy.NewProxyStore(server.ClientFactory, nil, server.AccessSetLookup, nil),
		vms:      scaled.VirtFactory.Kubevirt().V1().VirtualMachine(),
		vmCache:  scaled.VirtFactory.Kubevirt().V1().VirtualMachine().Cache(),
		pvcs:     scaled.CoreFactory.Core().V1().PersistentVolumeClaim(),
		pvcCache: scaled.CoreFactory.Core().V1().PersistentVolumeClaim().Cache(),
	}

	actionHandler := vmActionHandler{
		namespace:     options.Namespace,
		kubevirtCache: kubevirtCache,
		vms:           vms,
		vmis:          vmis,
		vmCache:       vms.Cache(),
		vmims:         vmims,
		vmimsc:        vmims.Cache(),
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
		// Formatter: vmformatter.formatter,
		Store: vmStore,
	}
	server.SchemaFactory.AddTemplate(t)

	return nil

}
