package stageverifier

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"io"
	"koality/verification"
	"koality/verification/config/commandgroup"
	"koality/vm"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type StageVerifier struct {
	virtualMachine vm.VirtualMachine
	ResultsChan    chan verification.Result
	changeStatus   *verification.ChangeStatus
}

func New(virtualMachine vm.VirtualMachine, changeStatus *verification.ChangeStatus) *StageVerifier {
	return &StageVerifier{
		virtualMachine: virtualMachine,
		ResultsChan:    make(chan verification.Result),
		changeStatus:   changeStatus,
	}
}

func (stageVerifier *StageVerifier) RunChangeStages(setupCommands commandgroup.CommandGroup, compileCommands commandgroup.CommandGroup, factoryCommands commandgroup.CommandGroup, testCommands commandgroup.AppendableCommandGroup) error {
	shouldContinue, err := stageVerifier.runSetupCommands(setupCommands)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	shouldContinue, err = stageVerifier.runCompileCommands(setupCommands)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	shouldContinue, err = stageVerifier.runFactoryCommands(factoryCommands, testCommands)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	factoryCommands.Wait()

	_, err = stageVerifier.runTestCommands(testCommands)
	return err
}

func (stageVerifier *StageVerifier) runSetupCommands(setupCommands commandgroup.CommandGroup) (bool, error) {
	var err error
	var setupCommand verification.Command
	for setupCommand, err = setupCommands.Next(); setupCommand != nil && err == nil; setupCommand, err = setupCommands.Next() {
		if stageVerifier.changeStatus.Cancelled || stageVerifier.changeStatus.Failed {
			return false, nil
		}
		returnCode, err := stageVerifier.runCommand(setupCommand, os.Stdout, os.Stderr)
		setupCommands.Done()
		stageVerifier.ResultsChan <- verification.Result{"setup", returnCode == 0 && err == nil}
		if err != nil {
			return false, err
		}
		if returnCode != 0 {
			return false, nil
		}
	}
	if err == commandgroup.NoMoreCommands {
		return true, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runCompileCommands(compileCommands commandgroup.CommandGroup) (bool, error) {
	var err error
	for compileCommand, err := compileCommands.Next(); compileCommand != nil && err == nil; compileCommand, err = compileCommands.Next() {
		if stageVerifier.changeStatus.Cancelled || stageVerifier.changeStatus.Failed {
			return false, nil
		}
		returnCode, err := stageVerifier.runCommand(compileCommand, os.Stdout, os.Stderr)
		compileCommands.Done()
		stageVerifier.ResultsChan <- verification.Result{"compile", returnCode == 0 && err == nil}
		if err != nil {
			return false, err
		}
		if returnCode != 0 {
			return false, nil
		}
	}
	if err == nil || err == commandgroup.NoMoreCommands {
		return true, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runFactoryCommands(factoryCommands commandgroup.CommandGroup, testCommands commandgroup.AppendableCommandGroup) (bool, error) {
	outputBuffer := new(bytes.Buffer)
	syncOutputBuffer := &syncWriter{writer: outputBuffer}
	var err error
	for factoryCommand, err := factoryCommands.Next(); factoryCommand != nil && err == nil; factoryCommand, err = factoryCommands.Next() {
		if stageVerifier.changeStatus.Cancelled || stageVerifier.changeStatus.Failed {
			return false, nil
		}
		returnCode, err := stageVerifier.runCommand(factoryCommand, io.MultiWriter(syncOutputBuffer, os.Stdout), io.MultiWriter(syncOutputBuffer, os.Stderr))
		factoryCommands.Done()
		stageVerifier.ResultsChan <- verification.Result{"factory", returnCode == 0 && err == nil}
		if err != nil {
			return false, err
		}
		if returnCode != 0 {
			return false, nil
		}
		// TODO (bbland): parse the output into new commands
		newTests := []verification.Command{}
		for _, newTest := range newTests {
			testCommands.Append(newTest)
		}
	}
	if err == nil || err == commandgroup.NoMoreCommands {
		return true, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runTestCommands(testCommands commandgroup.CommandGroup) (bool, error) {
	testsSuccess := true
	var err error
	for testCommand, err := testCommands.Next(); testCommand != nil && err == nil; testCommand, err = testCommands.Next() {
		if stageVerifier.changeStatus.Cancelled {
			return false, nil
		}
		returnCode, err := stageVerifier.runCommand(testCommand, os.Stdout, os.Stderr)
		testCommands.Done()
		stageVerifier.ResultsChan <- verification.Result{"test", returnCode == 0 && err == nil}
		testsSuccess = testsSuccess && returnCode == 0
		if err != nil {
			return false, err
		}
	}
	if err == nil || err == commandgroup.NoMoreCommands {
		return testsSuccess, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runCommand(command verification.Command, stdout io.Writer, stderr io.Writer) (int, error) {
	shellCommand := command.ShellCommand()
	executable, err := stageVerifier.virtualMachine.MakeExecutable(shellCommand, stdout, stderr)
	if err != nil {
		return -1, err
	}
	exitErr := executable.Run()
	if exitErr != nil {
		switch exitErr.(type) {
		case *ssh.ExitError:
			sshErr := exitErr.(*ssh.ExitError)
			return sshErr.Waitmsg.ExitStatus(), nil
		case *exec.ExitError:
			execErr := exitErr.(*exec.ExitError)
			// This only works for unix-type systems right now
			waitStatus, ok := execErr.Sys().(syscall.WaitStatus)
			if ok {
				return waitStatus.ExitStatus(), nil
			}
		}
		return -1, exitErr
	}
	return 0, nil
}

type syncWriter struct {
	writer io.Writer
	mutex  sync.Mutex
}

func (writer *syncWriter) Write(bytes []byte) (int, error) {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	return writer.writer.Write(bytes)
}
