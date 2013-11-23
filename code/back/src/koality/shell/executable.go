package shell

import (
	"io"
	"os/exec"
)

type Executable interface {
	SetStdout(io.Writer) error
	SetStderr(io.Writer) error
	Start() error
	Run() error
	Wait() error
}

type ExecutableMaker interface {
	MakeExecutable(Command) (Executable, error)
}

type ShellExecutableMaker struct{} // Empty struct?

type ShellExecutable struct {
	*exec.Cmd
}

func NewShellExecutableMaker() *ShellExecutableMaker {
	return &ShellExecutableMaker{}
}

func (executableMaker *ShellExecutableMaker) MakeExecutable(command Command) (Executable, error) {
	return ShellExecutable{exec.Command("bash", "-c", string(command))}, nil
}

func (executable ShellExecutable) SetStdout(writer io.Writer) error {
	executable.Cmd.Stdout = writer
	return nil
}

func (executable ShellExecutable) SetStderr(writer io.Writer) error {
	executable.Stderr = writer
	return nil
}

func (executable ShellExecutable) Start() error {
	return executable.Cmd.Start()
}

func (executable ShellExecutable) Run() error {
	return executable.Cmd.Run()
}

func (executable ShellExecutable) Wait() error {
	return executable.Cmd.Wait()
}
