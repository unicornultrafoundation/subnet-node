package main

import (
	"os"

	controllergen "github.com/rancher/wrangler/v3/pkg/controller-gen"
	"github.com/rancher/wrangler/v3/pkg/controller-gen/args"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

func main() {
	os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/generated",
		Boilerplate:   "core/kubevirt/scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			kubevirtv1.SchemeGroupVersion.Group: {
				Types: []interface{}{
					kubevirtv1.VirtualMachine{},
					kubevirtv1.VirtualMachineInstance{},
					kubevirtv1.VirtualMachineInstanceMigration{},
					kubevirtv1.KubeVirt{},
				},
				GenerateTypes:   false,
				GenerateClients: true,
			},
		},
	})
}
