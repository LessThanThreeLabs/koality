package shell

import (
	// "bytes"
	"io"
	// "io/ioutil"
	// "os"
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
	MakeExecutable(Command) Executable
}

type ShellExecutableMaker struct{} // Empty struct?

func NewShellExecutableMaker() *ShellExecutableMaker {
	return &ShellExecutableMaker{}
}

func (executableMaker *ShellExecutableMaker) MakeExecutable(command Command) Executable {
	return exec.Command("bash", "-c", string(command))
}
