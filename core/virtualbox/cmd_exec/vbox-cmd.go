package cmd_exec

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

var (
	VBoxCmd Command
)

// GetVBoxCmd returns the Command to run VBoxManage/VBoxControl.
func GetVBoxCmd() Command {
	if VBoxCmd != nil {
		return VBoxCmd
	}

	sudoer, err := isSudoer()
	if err != nil {
		return nil
	}

	if vbprog, err := lookupVBoxProgram("VBoxManage"); err == nil {
		VBoxCmd = command{program: vbprog, sudoer: sudoer, guest: false, cmdType: CommandTypeVBox}
	} else if vbprog, err := lookupVBoxProgram("VBoxControl"); err == nil {
		VBoxCmd = command{program: vbprog, sudoer: sudoer, guest: true, cmdType: CommandTypeVBox}
	} else {
		// Did not find a VirtualBox management command
		VBoxCmd = command{program: "false", sudoer: false, guest: false, cmdType: CommandTypeVBox}
	}
	// Debug("manage: '%+v'", manage)
	return VBoxCmd
}

// lookupVBoxProgram finds the path to the specified VirtualBox program (e.g., VBoxManage).
func lookupVBoxProgram(vbprog string) (string, error) {
	// Check VBOX_INSTALL_PATH environment variable first (applicable across platforms)
	if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" {
		if runtime.GOOS == "windows" {
			vbprog = filepath.Join(p, vbprog+".exe")
		} else {
			vbprog = filepath.Join(p, vbprog)
		}
		return exec.LookPath(vbprog)
	}

	// Platform-specific default paths
	switch runtime.GOOS {
	case "windows":
		vbprog = filepath.Join("C:\\", "Program Files", "Oracle", "VirtualBox", vbprog+".exe")
	case "darwin": // macOS
		vbprog = filepath.Join("/Applications/VirtualBox.app/Contents/MacOS", vbprog)
	case "linux": // Ubuntu
		vbprog = filepath.Join("/usr/bin", vbprog)
		// Also check /usr/local/bin as a fallback
		if _, err := exec.LookPath(vbprog); err != nil {
			vbprog = filepath.Join("/usr/local/bin", vbprog)
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return exec.LookPath(vbprog)
}

func isSudoer() (bool, error) {
	me, err := user.Current()
	if err != nil {
		return false, err
	}
	if groupIDs, err := me.GroupIds(); runtime.GOOS == "linux" {
		if err != nil {
			return false, err
		}
		for _, groupID := range groupIDs {
			group, err := user.LookupGroupId(groupID)
			if err != nil {
				return false, err
			}
			if group.Name == "sudo" {
				return true, nil
			}
		}
	}
	return false, nil
}
