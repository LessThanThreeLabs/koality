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
	Id() uint64
	Get(uint64) (<-chan VirtualMachine, <-chan error)
	Free()
	MaxSize() uint64
	SetMaxSize(uint64) error
	MinReady() uint64
	SetMinReady(uint64) error
}
