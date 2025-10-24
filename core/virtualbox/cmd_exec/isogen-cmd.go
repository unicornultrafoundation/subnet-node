package cmd_exec

import "runtime"

var (
	IsogenCmd Command
)

func GetIsogenCmd() Command {
	if IsogenCmd != nil {
		return IsogenCmd
	}
	// isogenimage if OS is Ubuntu, otherwise mkisofs
	if runtime.GOOS == osLinux {
		IsogenCmd = command{program: "isogenimage", sudoer: false, guest: false, cmdType: CommandTypeIsogen}
	} else {
		IsogenCmd = command{program: "mkisofs", sudoer: false, guest: false, cmdType: CommandTypeIsogen}
	}
	return IsogenCmd
}
