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
	GetN(uint64) <-chan VirtualMachine
	Free()
	MaxSize() uint64
	SetMaxSize(uint64)
	MinReady() uint64
	SetMinReady(uint64)
}
