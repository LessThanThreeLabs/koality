package stagerunner

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"encoding/json"
	"fmt"
	"github.com/dchest/goyaml"
	"io"
	"koality/resources"
	"koality/shell"
	"koality/verification"
	"koality/verification/config"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/vm"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"syscall"
	"time"
)

type StageRunner struct {
	ResultsChan         chan verification.SectionResult
	resourcesConnection *resources.Connection
	virtualMachine      vm.VirtualMachine
	verification        *resources.Verification
}

func New(resourcesConnection *resources.Connection, virtualMachine vm.VirtualMachine, currentVerification *resources.Verification) *StageRunner {
	return &StageRunner{
		ResultsChan:         make(chan verification.SectionResult),
		resourcesConnection: resourcesConnection,
		virtualMachine:      virtualMachine,
		verification:        currentVerification,
	}
}

func (stageRunner *StageRunner) RunStages(sections, finalSections []section.Section, environment map[string]string) error {
	for sectionNumber, section := range sections {
		shouldContinue, err := stageRunner.runSection(uint64(sectionNumber), section, environment)
		if err != nil {
			return err
		}
		if !shouldContinue {
			break
		}
	}
	if len(finalSections) == 0 {
		return nil
	}
	// TODO (bbland): do something smarter than just sleep and poll
	for stageRunner.verification.Status == "running" {
		time.Sleep(time.Second)
	}
	if stageRunner.verification.Status == "cancelled" {
		return nil
	}
	if environment == nil {
		environment = make(map[string]string)
	}
	environment["KOALITY_STATUS"] = stageRunner.verification.Status
	environment["KOALITY_MERGE_STATUS"] = stageRunner.verification.MergeStatus
	for finalSectionNumber, finalSection := range finalSections {
		shouldContinue, err := stageRunner.runSection(uint64(finalSectionNumber+len(sections)), finalSection, environment)
		if err != nil {
			return err
		}
		if !shouldContinue {
			break
		}
	}
	return nil
}

func (stageRunner *StageRunner) runSection(sectionNumber uint64, section section.Section, environment map[string]string) (bool, error) {
	factorySuccess, err := stageRunner.runFactoryCommands(sectionNumber, section, environment)
	if err != nil {
		return false, err
	}
	if !factorySuccess && !section.ContinueOnFailure() {
		return false, nil
	}

	commandsSuccess, err := stageRunner.runCommands(!factorySuccess, sectionNumber, section, environment)
	if err != nil {
		return false, err
	}

	if factorySuccess && commandsSuccess {
		stageRunner.ResultsChan <- verification.SectionResult{
			Section:       section.Name(),
			Final:         section.IsFinal(),
			FailSectionOn: section.FailOn(),
			Passed:        true,
		}
	}
	return commandsSuccess || section.ContinueOnFailure(), nil
}

func (stageRunner *StageRunner) getXunitResults(command verification.Command, environment map[string]string) ([]resources.XunitResult, error) {
	execName := "getXunitResults"
	directories := command.XunitPaths()
	if len(directories) == 0 {
		return nil, nil
	}
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	localPath := path.Join(usr.HomeDir, "code", "back", "src", "koality", "util", execName)
	copyExec, err := stageRunner.virtualMachine.FileCopy(localPath, execName)
	if err != nil {
		return nil, err
	}

	if err = copyExec.Run(); err != nil {
		return nil, err
	}

	cmd := shell.Commandf("./%s %s %s", execName, shell.Quote("*.xml"), strings.Join(directories, " "))
	w := new(bytes.Buffer)
	runExec, err := stageRunner.virtualMachine.MakeExecutable(cmd, nil, w, os.Stderr, environment)
	if err = runExec.Run(); err != nil {
		return nil, err
	}

	var res []resources.XunitResult
	if err = json.Unmarshal(w.Bytes(), &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (stageRunner *StageRunner) runFactoryCommands(sectionNumber uint64, sectionToRun section.Section, environment map[string]string) (bool, error) {
	var err error
	var command verification.Command
	sectionFailed := false

	factoryCommands := sectionToRun.FactoryCommands(false)

	for command, err = factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
		if stageRunner.verification.Status == "cancelled" {
			return false, nil
		}

		stages, err := stageRunner.resourcesConnection.Stages.Read.GetAll(stageRunner.verification.Id)
		if err != nil {
			return false, err
		}

		var stageId uint64

		for _, stage := range stages {
			if stage.SectionNumber == sectionNumber && stage.Name == command.Name() {
				stageId = stage.Id
				break
			}
		}
		if stageId == 0 {
			return false, fmt.Errorf("Unable to find a stage to match %#v", sectionToRun)
		}

		stageRun, err := stageRunner.resourcesConnection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return false, err
		}

		outputBuffer := new(bytes.Buffer)
		syncOutputBuffer := &syncWriter{writer: outputBuffer}

		consoleWriter := newConsoleTextWriter(stageRunner.resourcesConnection.Stages.Update, stageRun.Id)
		syncConsoleWriter := &syncWriter{writer: consoleWriter}

		err = stageRunner.resourcesConnection.Stages.Update.SetStartTime(stageRun.Id, time.Now())
		if err != nil {
			return false, err
		}

		returnCode, runErr := stageRunner.runCommand(command, nil, io.MultiWriter(syncOutputBuffer, syncConsoleWriter, os.Stdout), io.MultiWriter(syncOutputBuffer, syncConsoleWriter, os.Stderr), environment)
		factoryCommands.Done()
		closeErr := consoleWriter.Close()

		stageFailed := returnCode != 0 || runErr != nil
		if stageFailed && !sectionFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				Final:         sectionToRun.IsFinal(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}

		setEndTimeErr := stageRunner.resourcesConnection.Stages.Update.SetEndTime(stageRun.Id, time.Now())
		setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRun.Id, returnCode)

		if closeErr != nil {
			return false, closeErr
		}
		if runErr != nil {
			return false, runErr
		}
		if setEndTimeErr != nil {
			return false, setEndTimeErr
		}
		if setReturnCodeErr != nil {
			return false, setReturnCodeErr
		}

		if stageFailed && !sectionToRun.ContinueOnFailure() {
			return false, nil
		}
		var yamlParsedCommands interface{}

		err = goyaml.Unmarshal(outputBuffer.Bytes(), &yamlParsedCommands)
		if err != nil {
			// TODO (bbland): display an error to the user
			return false, err
		}
		newCommands, err := config.ParseRemoteCommands(yamlParsedCommands, true)
		if err != nil {
			// TODO (bbland): display the error to the user
			return false, err
		}
		stages, err = stageRunner.resourcesConnection.Stages.Read.GetAll(stageRunner.verification.Id)
		if err != nil {
			return false, err
		}
		var maxOrderNumber uint64
		for _, stage := range stages {
			if stage.SectionNumber == sectionNumber && stage.OrderNumber > maxOrderNumber {
				maxOrderNumber = stage.OrderNumber
			}
		}
		orderNumber := maxOrderNumber + 1
		for _, newCommand := range newCommands {
			// This can still end up with duplicate/bad order numbers if multiple factories run simultaneously
			command, err := sectionToRun.AppendCommand(newCommand)
			if err != nil {
				return false, err
			}

			_, err = stageRunner.resourcesConnection.Stages.Create.Create(stageRunner.verification.Id, sectionNumber, command.Name(), orderNumber)
			if err != nil {
				return false, err
			}

			orderNumber++
		}
	}
	if err != nil && err != commandgroup.NoMoreCommands {
		return false, err
	}
	return !sectionFailed, nil
}

func (stageRunner *StageRunner) runCommands(sectionPreviouslyFailed bool, sectionNumber uint64, sectionToRun section.Section, environment map[string]string) (bool, error) {
	var err error
	var command verification.Command
	sectionFailed := false

	commands := sectionToRun.Commands(false)
	for command, err = commands.Next(); err == nil; command, err = commands.Next() {
		if stageRunner.verification.Status == "cancelled" {
			return false, nil
		}

		stages, err := stageRunner.resourcesConnection.Stages.Read.GetAll(stageRunner.verification.Id)
		if err != nil {
			return false, err
		}

		var stageId uint64

		for _, stage := range stages {
			if stage.SectionNumber == sectionNumber && stage.Name == command.Name() {
				stageId = stage.Id
				break
			}
		}
		if stageId == 0 {
			return false, fmt.Errorf("Unable to find a stage to match %#v", sectionToRun)
		}

		stageRun, err := stageRunner.resourcesConnection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return false, err
		}

		consoleWriter := newConsoleTextWriter(stageRunner.resourcesConnection.Stages.Update, stageRun.Id)
		syncConsoleWriter := &syncWriter{writer: consoleWriter}

		err = stageRunner.resourcesConnection.Stages.Update.SetStartTime(stageRun.Id, time.Now())
		if err != nil {
			return false, err
		}

		returnCode, runErr := stageRunner.runCommand(command, nil, io.MultiWriter(syncConsoleWriter, os.Stdout), io.MultiWriter(syncConsoleWriter, os.Stderr), environment)
		commands.Done()
		closeErr := consoleWriter.Close()

		stageFailed := returnCode != 0 || runErr != nil
		if stageFailed && !sectionFailed && !sectionPreviouslyFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				Final:         sectionToRun.IsFinal(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}

		setEndTimeErr := stageRunner.resourcesConnection.Stages.Update.SetEndTime(stageRun.Id, time.Now())
		setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRun.Id, returnCode)
		xunitResults, xunitError := stageRunner.getXunitResults(command, environment)
		if xunitError != nil {
			return false, xunitError
		}

		if len(xunitResults) > 0 {
			xunitError = stageRunner.resourcesConnection.Stages.Update.AddXunitResults(stageRun.Id, xunitResults)
			if xunitError != nil {
				return false, xunitError
			}
		}

		if closeErr != nil {
			return false, closeErr
		}
		if runErr != nil {
			return false, runErr
		}
		if setEndTimeErr != nil {
			return false, setEndTimeErr
		}
		if setReturnCodeErr != nil {
			return false, setReturnCodeErr
		}

		if stageFailed && !sectionToRun.ContinueOnFailure() {
			return false, nil
		}
	}
	if err != nil && err != commandgroup.NoMoreCommands {
		return false, err
	}
	return !sectionFailed, nil
}

func (stageRunner *StageRunner) runCommand(command verification.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (int, error) {
	shellCommand := command.ShellCommand()
	executable, err := stageRunner.virtualMachine.MakeExecutable(shellCommand, stdin, stdout, stderr, environment)
	if err != nil {
		return 255, err
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
		return 255, exitErr
	}
	return 0, nil
}
