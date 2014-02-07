package localmachine

import (
	"io"
	"io/ioutil"
	"koality/shell"
	"koality/vm"
	"os"
	"path"
)

type LocalMachine struct {
	rootDir         string
	executableMaker shell.ExecutableMaker
	copier          *localCopier
	patcher         vm.Patcher
}

func New() (*LocalMachine, error) {
	rootDir, err := ioutil.TempDir("", "fakevm-")
	if err != nil {
		return nil, err
	}
	return FromDir(rootDir)
}

func FromDir(rootDir string) (*LocalMachine, error) {
	executableMaker := shell.ShellExecutableMaker
	copier := &localCopier{executableMaker}

	localMachine := LocalMachine{
		rootDir:         rootDir,
		executableMaker: executableMaker,
		copier:          copier,
		patcher:         vm.NewPatcher(copier, executableMaker),
	}

	setupExec, err := executableMaker.MakeExecutable(shell.Advertised(shell.Commandf("mkdir -p %s", rootDir)), nil, nil, nil, nil)
	if err != nil {
		os.RemoveAll(rootDir)
		return nil, err
	}
	if err = setupExec.Run(); err != nil {
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

func (localMachine *LocalMachine) MakeExecutable(command shell.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (shell.Executable, error) {
	fullCommand := shell.And(
		shell.Commandf("cd %s", localMachine.rootDir),
		command,
	)
	if environment == nil {
		environment = make(map[string]string)
	}
	environment["HOME"] = localMachine.rootDir
	return localMachine.executableMaker.MakeExecutable(fullCommand, stdin, stdout, stderr, environment)
}

func (localMachine *LocalMachine) Patch(patchConfig *vm.PatchConfig) (shell.Executable, error) {
	return localMachine.patcher.Patch(patchConfig)
}

func (localMachine *LocalMachine) FileCopy(sourceFilePath, destFilePath string) (shell.Executable, error) {
	destPath := path.Join(localMachine.rootDir, destFilePath)
	if destFilePath[len(destFilePath)-1] == '/' {
		destPath += "/"
	}
	return localMachine.copier.FileCopy(sourceFilePath, destPath)
}

func (localMachine *LocalMachine) Terminate() error {
	return os.RemoveAll(localMachine.rootDir)
}

type localCopier struct {
	executableMaker shell.ExecutableMaker
}

func (copier *localCopier) FileCopy(sourceFilePath, destFilePath string) (shell.Executable, error) {
	command := shell.Advertised(shell.And(shell.Commandf("mkdir -p %s", path.Dir(destFilePath)),
		shell.Commandf("cp %s %s", sourceFilePath, destFilePath)))
	return copier.executableMaker.MakeExecutable(command, nil, nil, nil, nil)
}
