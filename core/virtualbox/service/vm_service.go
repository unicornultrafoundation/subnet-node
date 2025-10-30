package vbox_service

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/cmd_exec"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func (s *VBoxService) GetVM(vmId string) (*vbtypes.VM, error) {
	s.lock.Lock()

	stdout, stderr, err := vBoxCmd.Run("showvminfo", vmId, "--machinereadable")
	s.lock.Unlock()
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return nil, cmd_exec.ErrMachineNotExist
		}
		return nil, err
	}

	/* Read all VM info into a map */
	props := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		res := reVMInfoLine.FindStringSubmatch(scanner.Text())
		if res == nil {
			continue
		}
		key := res[1]
		if key == "" {
			key = res[2]
		}
		val := res[3]
		if val == "" {
			val = res[4]
		}
		props[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading machine info: %w", err)
	}

	// error that occured during parsing
	var perr error

	sp := func(field string, def ...string) string {
		if v, exists := props[field]; exists {
			return v
		}
		if len(def) < 1 {
			return ""
		}
		return def[0]
	}

	up := func(field string, def ...uint) uint {
		if v, exists := props[field]; exists {
			n, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				perr = err
				return 0
			}
			return uint(n)
		}
		if len(def) < 1 {
			return 0
		}
		return def[0]
	}

	/* Extract basic info */
	vm := &vbtypes.VM{
		ID:         sp("UUID"),
		Name:       sp("name"),
		Status:     vbtypes.VMStatus(sp("VMState")),
		CPUCores:   int(up("cpus")),
		MemoryMB:   int(up("memory")),
		DiskSizeGB: int(up("storagebytes")),
		// IPAddress: sp("IPAddress"),
		// SSHPort: up("SSHPort"),
		VMFolder: sp("CfgFile"),
	}

	if perr != nil {
		return nil, fmt.Errorf("parsing machine props failed: %w", perr)
	}

	return vm, nil
}

func (s *VBoxService) ListVMs() ([]*vbtypes.VM, error) {

	stdout, _, err := vBoxCmd.Run("list", "vms")
	if err != nil {
		return nil, fmt.Errorf("unable to list vms: %w", err)
	}
	vms := []*vbtypes.VM{}
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		res := reVMNameUUID.FindStringSubmatch(scanner.Text())
		if res == nil {
			continue
		}
		m, err := s.GetVM(res[1])
		if err != nil {
			// Sometimes a VM is listed but not available, so we need to handle this.
			if errors.Is(err, cmd_exec.ErrMachineNotExist) {
				continue
			} else {
				return nil, fmt.Errorf("unable to get machine info: %w", err)
			}
		}
		vms = append(vms, m)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading machine list: %w", err)
	}
	return vms, nil
}

func (s *VBoxService) createVM(name string) (*vbtypes.VM, error) {

	if name == "" {
		return nil, fmt.Errorf("machine name is empty")
	}

	// Check if a machine with the given name already exists.
	ms, err := s.ListVMs()
	if err != nil {
		return nil, err
	}
	for _, m := range ms {
		if m.Name == name {
			return nil, cmd_exec.ErrMachineExist
		}
	}

	// Create and register the machine.
	args := []string{"createvm", "--name", name, "--register"}

	if _, _, err = vBoxCmd.Run(args...); err != nil {
		return nil, err
	}

	return s.GetVM(name)
}
