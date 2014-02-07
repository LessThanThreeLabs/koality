package localmachine

import (
	"io/ioutil"
	"koality/shell"
	"koality/vm"
	"os"
)

type LocalMachine struct {
	rootDir  string
	executor shell.Executor
}

func New() (*LocalMachine, error) {
	rootDir, err := ioutil.TempDir("", "fakevm-")
	if err != nil {
		return nil, err
	}
	return FromDir(rootDir)
}

func FromDir(rootDir string) (*LocalMachine, error) {
	localMachine := LocalMachine{
		rootDir:  rootDir,
		executor: shell.ShellExecutor,
	}

	setupExecutable := shell.Executable{
		Command: shell.Commandf("mkdir -p %s", shell.Quote(rootDir)),
	}

	setupExecution, err := localMachine.Execute(setupExecutable)
	if err != nil {
		os.RemoveAll(rootDir)
		return nil, err
	}
	if err = setupExecution.Wait(); err != nil {
		os.RemoveAll(rootDir)
		return nil, err
	}

	return &localMachine, nil
}

func (localMachine LocalMachine) Id() string {
	return "localhost"
}

func (localMachine LocalMachine) GetStartShellCommand() vm.Command {
	return vm.Command{Argv: []string{
		"/bin/bash",
		"/bin/bash -c " + shell.Quote("cd "+localMachine.rootDir+" && /bin/bash"),
	}}
}

func (localMachine *LocalMachine) Execute(executable shell.Executable) (shell.Execution, error) {
	executable.Command = shell.And(
		shell.Commandf("cd %s", localMachine.rootDir),
		executable.Command,
	)

	if executable.Environment == nil {
		executable.Environment = make(map[string]string)
	}
	executable.Environment["HOME"] = localMachine.rootDir

	return localMachine.executor.Execute(executable)
}

func (localMachine *LocalMachine) Terminate() error {
	return os.RemoveAll(localMachine.rootDir)
}
