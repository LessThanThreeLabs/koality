package localmachine

import (
	"fmt"
	"io"
	"io/ioutil"
	"koality/shell"
	"koality/vm"
	"os"
	"path/filepath"
)

type LocalMachine struct {
	rootDir         string
	executableMaker shell.ExecutableMaker
	copier          *localCopier
	patcher         vm.Patcher
}

func New() *LocalMachine {
	rootDir, err := ioutil.TempDir("", "fakevm-")
	if err != nil {
		panic(err)
	}
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
		panic(err)
	}
	setupExec.Run()

	return &localMachine
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
	return localMachine.copier.FileCopy(sourceFilePath, fmt.Sprintf("%s/%s", localMachine.rootDir, destFilePath))
}

func (localMachine *LocalMachine) Terminate() error {
	return os.RemoveAll(localMachine.rootDir)
}

type localCopier struct {
	executableMaker shell.ExecutableMaker
}

func (copier *localCopier) FileCopy(sourceFilePath, destFilePath string) (shell.Executable, error) {
	command := shell.Advertised(shell.And(shell.Commandf("mkdir -p %s", filepath.Dir(destFilePath)),
		shell.Commandf("cp %s %s", sourceFilePath, destFilePath)))
	return copier.executableMaker.MakeExecutable(command, nil, nil, nil, nil)
}
