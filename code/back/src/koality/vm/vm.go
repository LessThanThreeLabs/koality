package vm

import (
	"koality/shell"
        "syscall"
)

type VirtualMachine interface {
	shell.ExecutableMaker
	Patcher
	shell.FileCopier

	GetStartShellCommand() Command
	Terminate() error
}

type Command struct {
	Argv []string
	Envv []string
}

type VirtualMachineManager interface {
	NewVirtualMachine() (VirtualMachine, error)
	GetVirtualMachine(string) (VirtualMachine, error)
}

type VirtualMachinePool interface {
	Id() uint64
	GetReady(uint64) (<-chan VirtualMachine, <-chan error)
	GetExisting(string) (VirtualMachine, error)
	Free()
	Return(VirtualMachine)
	MaxSize() uint64
	SetMaxSize(uint64) error
	MinReady() uint64
	SetMinReady(uint64) error
}

func (command Command) Exec() error {
	return syscall.Exec(command.Argv[0], command.Argv, command.Envv)
}
