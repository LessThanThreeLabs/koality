package vm

import (
	"koality/shell"
)

type VirtualMachine interface {
	shell.ExecutableMaker
	Patcher
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
	MinReady() int
	SetMinReady(int)
}
