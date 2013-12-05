package stagerunner

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"io"
	"koality/resources"
	"koality/verification"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/vm"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type StageRunner struct {
	virtualMachine vm.VirtualMachine
	ResultsChan    chan verification.SectionResult
	verification   *resources.Verification
}

func New(virtualMachine vm.VirtualMachine, currentVerification *resources.Verification) *StageRunner {
	return &StageRunner{
		virtualMachine: virtualMachine,
		ResultsChan:    make(chan verification.SectionResult),
		verification:   currentVerification,
	}
}

func (stageRunner *StageRunner) RunStages(sections []section.Section) error {
	for _, section := range sections {
		shouldContinue, err := stageRunner.runSection(section)
		if err != nil || !shouldContinue {
			return err
		}
	}
	return nil
}

func (stageRunner *StageRunner) runSection(section section.Section) (bool, error) {
	factorySuccess, err := stageRunner.runFactoryCommands(section)
	if err != nil {
		return false, err
	}
	if !factorySuccess && !section.ContinueOnFailure() {
		return false, nil
	}
	commandsSuccess, err := stageRunner.runCommands(!factorySuccess, section)
	if err != nil {
		return false, err
	}
	if factorySuccess && commandsSuccess {
		stageRunner.ResultsChan <- verification.SectionResult{
			Section:       section.Name(),
			FailSectionOn: section.FailOn(),
			Passed:        true,
		}
	}
	return commandsSuccess || section.ContinueOnFailure(), nil
}

func (stageRunner *StageRunner) runFactoryCommands(sectionToRun section.Section) (bool, error) {
	var err error
	var command verification.Command
	sectionFailed := false

	factoryCommands := sectionToRun.FactoryCommands()
	for command, err = factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
		if stageRunner.verification.VerificationStatus == "cancelled" {
			return false, nil
		}
		outputBuffer := new(bytes.Buffer)
		syncOutputBuffer := &syncWriter{writer: outputBuffer}
		returnCode, err := stageRunner.runCommand(command, os.Stdin, io.MultiWriter(syncOutputBuffer, os.Stdout), io.MultiWriter(syncOutputBuffer, os.Stderr))
		factoryCommands.Done()
		stageFailed := returnCode != 0 || err != nil
		if stageFailed && !sectionFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}
		if err != nil {
			return false, err
		}
		if stageFailed && !sectionToRun.ContinueOnFailure() {
			return false, nil
		}
		// TODO (bbland): parse the output into new commands
		newCommands := []verification.Command{}
		for _, newCommand := range newCommands {
			sectionToRun.AppendCommand(newCommand)
		}
	}
	if err != commandgroup.NoMoreCommands {
		return false, err
	}
	return !sectionFailed, nil
}

func (stageRunner *StageRunner) runCommands(sectionPreviouslyFailed bool, sectionToRun section.Section) (bool, error) {
	var err error
	var command verification.Command
	sectionFailed := false

	commands := sectionToRun.Commands()
	for command, err = commands.Next(); err == nil; command, err = commands.Next() {
		if stageRunner.verification.VerificationStatus == "cancelled" {
			return false, nil
		}
		returnCode, err := stageRunner.runCommand(command, os.Stdin, os.Stdout, os.Stderr)
		commands.Done()
		stageFailed := returnCode != 0 || err != nil
		if stageFailed && !sectionFailed && !sectionPreviouslyFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}
		if err != nil {
			return false, err
		}
		if stageFailed && !sectionToRun.ContinueOnFailure() {
			return false, nil
		}
	}
	if err != commandgroup.NoMoreCommands {
		return false, err
	}
	return sectionFailed, nil
}

// func (stageRunner *StageRunner) RunStages(setupCommands commandgroup.CommandGroup, firstPreTestCommands commandgroup.CommandGroup, everyPreTestCommands commandgroup.CommandGroup, factoryCommands commandgroup.CommandGroup, testCommands commandgroup.AppendableCommandGroup, everyPostTestCommands commandgroup.CommandGroup, lastPostTestCommands commandgroup.CommandGroup) error {
// 	shouldContinue, err := stageRunner.runSetupCommands(setupCommands)
// 	if err != nil || !shouldContinue {
// 		return err
// 	}
// 	shouldContine, err = stageRunner.runFirstPreTestCommands(firstPreTestCommands)
// 	if err != nil || !shouldContinue {
// 		return err
// 	}
// 	shouldContinue, err = stageRunner.runEveryPreTestCommands(everyPreTestCommands)
// 	if err != nil || !shouldContinue {
// 		return err
// 	}
// 	shouldContinue, err = stageRunner.runFactoryCommands(factoryCommands, testCommands)
// 	if err != nil || !shouldContinue {
// 		return err
// 	}
// 	factoryCommands.Wait()

// 	_, err = stageRunner.runTestCommands(testCommands)
// 	if err != nil {
// 		return err
// 	}
// 	shouldContinue, err = stageRunner.runEveryPostTestCommands(everyPostTestCommands)
// 	if err != nil || !shouldContinue {
// 		return err
// 	}
// 	shouldContinue, err = stageRunner.runLastPostTestCommands(lastPreTestCommands)
// 	if err != nil || !shouldContinue {
// 		return err
// 	}
// 	// Do other stuff?
// 	return err
// }

// func (stageRunner *StageRunner) runSetupCommands(setupCommands commandgroup.CommandGroup) (bool, error) {
// 	var err error
// 	var setupCommand verification.Command
// 	for setupCommand, err = setupCommands.Next(); setupCommand != nil && err == nil; setupCommand, err = setupCommands.Next() {
// 		if stageRunner.verification.VerificationStatus == "cancelled" || stageRunner.verification.VerificationStatus == "failed" {
// 			return false, nil
// 		}
// 		returnCode, err := stageRunner.runCommand(setupCommand, os.Stdin, os.Stdout, os.Stderr)
// 		setupCommands.Done()
// 		stageRunner.ResultsChan <- verification.Result{"setup", returnCode == 0 && err == nil}
// 		if err != nil {
// 			return false, err
// 		}
// 		if returnCode != 0 {
// 			return false, nil
// 		}
// 	}
// 	if err == commandgroup.NoMoreCommands {
// 		return true, nil
// 	}
// 	return false, err
// }

// func (stageRunner *StageRunner) runCompileCommands(compileCommands commandgroup.CommandGroup) (bool, error) {
// 	var err error
// 	for compileCommand, err := compileCommands.Next(); compileCommand != nil && err == nil; compileCommand, err = compileCommands.Next() {
// 		if stageRunner.verification.VerificationStatus == "cancelled" || stageRunner.verification.VerificationStatus == "failed" {
// 			return false, nil
// 		}
// 		returnCode, err := stageRunner.runCommand(compileCommand, os.Stdin, os.Stdout, os.Stderr)
// 		compileCommands.Done()
// 		stageRunner.ResultsChan <- verification.Result{"compile", returnCode == 0 && err == nil}
// 		if err != nil {
// 			return false, err
// 		}
// 		if returnCode != 0 {
// 			return false, nil
// 		}
// 	}
// 	if err == nil || err == commandgroup.NoMoreCommands {
// 		return true, nil
// 	}
// 	return false, err
// }

// func (stageRunner *StageRunner) runFactoryCommands(factoryCommands commandgroup.CommandGroup, testCommands commandgroup.AppendableCommandGroup) (bool, error) {
// 	outputBuffer := new(bytes.Buffer)
// 	syncOutputBuffer := &syncWriter{writer: outputBuffer}
// 	var err error
// 	for factoryCommand, err := factoryCommands.Next(); factoryCommand != nil && err == nil; factoryCommand, err = factoryCommands.Next() {
// 		if stageRunner.verification.VerificationStatus == "cancelled" || stageRunner.verification.VerificationStatus == "failed" {
// 			return false, nil
// 		}
// 		returnCode, err := stageRunner.runCommand(factoryCommand, os.Stdin, io.MultiWriter(syncOutputBuffer, os.Stdout), io.MultiWriter(syncOutputBuffer, os.Stderr))
// 		factoryCommands.Done()
// 		stageRunner.ResultsChan <- verification.Result{"factory", returnCode == 0 && err == nil}
// 		if err != nil {
// 			return false, err
// 		}
// 		if returnCode != 0 {
// 			return false, nil
// 		}
// 		// TODO (bbland): parse the output into new commands
// 		newTests := []verification.Command{}
// 		for _, newTest := range newTests {
// 			testCommands.Append(newTest)
// 		}
// 	}
// 	if err == nil || err == commandgroup.NoMoreCommands {
// 		return true, nil
// 	}
// 	return false, err
// }

// func (stageRunner *StageRunner) runTestCommands(testCommands commandgroup.CommandGroup) (bool, error) {
// 	testsSuccess := true
// 	var err error
// 	for testCommand, err := testCommands.Next(); testCommand != nil && err == nil; testCommand, err = testCommands.Next() {
// 		if stageRunner.verification.VerificationStatus == "cancelled" {
// 			return false, nil
// 		}
// 		returnCode, err := stageRunner.runCommand(testCommand, os.Stdin, os.Stdout, os.Stderr)
// 		testCommands.Done()
// 		stageRunner.ResultsChan <- verification.Result{"test", returnCode == 0 && err == nil}
// 		testsSuccess = testsSuccess && returnCode == 0
// 		if err != nil {
// 			return false, err
// 		}
// 	}
// 	if err == nil || err == commandgroup.NoMoreCommands {
// 		return testsSuccess, nil
// 	}
// 	return false, err
// }

func (stageRunner *StageRunner) runCommand(command verification.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	shellCommand := command.ShellCommand()
	executable, err := stageRunner.virtualMachine.MakeExecutable(shellCommand, stdin, stdout, stderr)
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
