package localmachine

import (
	"fmt"
	"io/ioutil"
	"koality/shell"
	"koality/vm"
	"os"
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
	executableMaker := shell.NewShellExecutableMaker()
	copier := &localCopier{executableMaker}

	localMachine := LocalMachine{
		rootDir:         rootDir,
		executableMaker: executableMaker,
		copier:          copier,
		patcher:         vm.NewPatcher(copier, executableMaker),
	}

	cmdStr := fmt.Sprintf("mkdir -p %s", rootDir)
	setupExec, err := executableMaker.MakeExecutable(shell.Advertised(shell.Command(cmdStr)))
	if err != nil {
		panic(err)
	}
	setupExec.Run()

	return &localMachine
}

func (localMachine *LocalMachine) MakeExecutable(command shell.Command) (shell.Executable, error) {
	fullCommand := shell.And(
		shell.Command(fmt.Sprintf("cd %s", localMachine.rootDir)),
		command,
	)
	return localMachine.executableMaker.MakeExecutable(fullCommand)
}

func (localMachine *LocalMachine) ProvisionCommand() shell.Command {
	panic("Not implemented")
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
	command := shell.Advertised(shell.Command(fmt.Sprintf("cp %s %s", sourceFilePath, destFilePath)))
	return copier.executableMaker.MakeExecutable(command)
}
