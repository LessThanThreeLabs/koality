package vm

import (
	"koality/shell"
	"os/exec"
	"path/filepath"
	"syscall"
)

type VirtualMachine interface {
	shell.Executor

	Id() string
	GetStartShellCommand() Command
	SaveState(name string) (imageId string, err error)
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
	execPath, err := exec.LookPath(command.Argv[0])
	if err != nil {
		return err
	}

	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return err
	}

	return syscall.Exec(absExecPath, command.Argv, command.Envv)
}
