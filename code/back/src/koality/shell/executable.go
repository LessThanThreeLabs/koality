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
	MakeExecutable(command Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (Executable, error)
}

type ShellExecutableMaker struct{} // Empty struct?

type ShellExecutable struct {
	*exec.Cmd
}

func NewShellExecutableMaker() *ShellExecutableMaker {
	return &ShellExecutableMaker{}
}

func (executableMaker *ShellExecutableMaker) MakeExecutable(command Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (Executable, error) {
	cmd := exec.Command("bash", "-c", string(command))
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	env := make([]string, len(environment))
	index := 0

	for key, value := range environment {
		env[index] = key + "=" + value
		index++
	}
	cmd.Env = env

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
