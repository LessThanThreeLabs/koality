package shell

import (
	"io"
	"os"
	"os/exec"
)

type Executable struct {
	Command     Command
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Environment map[string]string
}

type Executor interface {
	Execute(executable Executable) (Execution, error)
}

type Execution interface {
	Wait() error
}

type shellExecutor struct{} // Empty struct constant?
var ShellExecutor = shellExecutor{}

func (executor shellExecutor) Execute(executable Executable) (Execution, error) {
	cmd := exec.Command("bash", "-c", string(executable.Command))
	cmd.Stdin = executable.Stdin
	cmd.Stdout = executable.Stdout
	cmd.Stderr = executable.Stderr

	env := make([]string, 0, len(executable.Environment))
	for key, value := range executable.Environment {
		env = append(env, key+"="+value)
	}
	cmd.Env = append(os.Environ(), env...)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}
