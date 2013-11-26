package shell

import (
	"io"
	"os/exec"
)

type Executable interface {
	Start() error
	Run() error
	Wait() error
}

type ExecutableMaker interface {
	MakeExecutable(command Command, stdout io.Writer, stderr io.Writer) (Executable, error)
}

type ShellExecutableMaker struct{} // Empty struct?

type ShellExecutable struct {
	*exec.Cmd
}

func NewShellExecutableMaker() *ShellExecutableMaker {
	return &ShellExecutableMaker{}
}

func (executableMaker *ShellExecutableMaker) MakeExecutable(command Command, stdout io.Writer, stderr io.Writer) (Executable, error) {
	cmd := exec.Command("bash", "-c", string(command))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return ShellExecutable{cmd}, nil
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
