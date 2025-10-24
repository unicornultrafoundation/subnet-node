package cmd_exec

var (
	QemuCmd Command
)

func GetQemuCmd() Command {
	if QemuCmd != nil {
		return QemuCmd
	}

	QemuCmd = command{program: "qemu-img", sudoer: false, guest: false, cmdType: CommandTypeQEMU}
	return QemuCmd
}
