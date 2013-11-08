package shell

import (
	"io"
	"os/exec"
)

type Executable interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Run() error
	Wait() error
}

type ExecutableMaker interface {
	MakeExecutable(Command) (Executable, error)
}

type ShellExecutableMaker struct{} // Empty struct?

func NewShellExecutableMaker() *ShellExecutableMaker {
	return &ShellExecutableMaker{}
}

func (executableMaker *ShellExecutableMaker) MakeExecutable(command Command) (Executable, error) {
	return exec.Command("bash", "-c", string(command)), nil
}
