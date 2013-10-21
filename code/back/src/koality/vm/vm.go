package vm

import (
	"koality/shell"
)

// TODO: move this elsewhere
type Provisioner interface {
	ProvisionCommand() shell.Command
}

type VirtualMachine interface {
	shell.ExecutableMaker
	Patcher
	Provisioner

	FileCopy(sourcePath, destPath string) shell.Executable
	Terminate() error
}

type VirtualMachineLauncher interface {
	LaunchVirtualMachine() VirtualMachine
}

type VirtualMachinePool interface {
	Get() VirtualMachine
	MaxSize() int
	SetMaxSize(int)
	MinSize() int
	SetMinSize(int)
}
