package cmd_exec

const (
	stringYes string = "Yes"
	osWindows string = "windows"
	osLinux   string = "linux"
	osDarwin  string = "darwin"
)

// CommandType represents the type of command executor
type CommandType int

const (
	CommandTypeQEMU CommandType = iota
	CommandTypeVBox
	CommandTypeIsogen
)

func (ct CommandType) String() string {
	switch ct {
	case CommandTypeQEMU:
		return "QEMU"
	case CommandTypeVBox:
		return "VBoxManage"
	case CommandTypeIsogen:
		return "Isogen"
	default:
		return "Unknown"
	}
}
