package vm

import (
	"log"

	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/builder"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

type VMHandler struct {
	vm *kubevirtv1.VirtualMachine
}

func NewVMHandler() *VMHandler {

	vmBuilder := builder.NewVMBuilder("")

	vm, err := vmBuilder.VM()
	if err != nil {
		log.Fatalf("Failed to build VM: %v", err)
	}

	return &VMHandler{vm: vm}
}
