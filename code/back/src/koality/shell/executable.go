package shell

import (
	"io"
	"os"
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

type shellExecutableMaker struct{} // Empty struct constant?
var ShellExecutableMaker = shellExecutableMaker{}

type ShellExecutable struct {
	*exec.Cmd
}

func (executableMaker shellExecutableMaker) MakeExecutable(command Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (Executable, error) {
	cmd := exec.Command("bash", "-c", string(command))
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	env := make([]string, 0, len(environment))
	for key, value := range environment {
		env = append(env, key+"="+value)
	}
	cmd.Env = append(os.Environ(), env...)

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
