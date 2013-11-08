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
	shell.FileCopier

	Terminate() error
}

type VirtualMachineLauncher interface {
	LaunchVirtualMachine() (VirtualMachine, error)
}

type VirtualMachinePool interface {
	Get() VirtualMachine
	GetN(int) <-chan VirtualMachine
	Free()
	MaxSize() int
	SetMaxSize(int)
	MinSize() int
	SetMinSize(int)
}
