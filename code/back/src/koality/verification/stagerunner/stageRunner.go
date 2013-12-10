package stagerunner

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"github.com/dchest/goyaml"
	"io"
	"koality/resources"
	"koality/verification"
	"koality/verification/config"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/vm"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type StageRunner struct {
	resourcesConnection *resources.Connection
	virtualMachine      vm.VirtualMachine
	ResultsChan         chan verification.SectionResult
	verification        *resources.Verification
}

func New(resourcesConnection *resources.Connection, virtualMachine vm.VirtualMachine, currentVerification *resources.Verification) *StageRunner {
	return &StageRunner{
		resourcesConnection: resourcesConnection,
		virtualMachine:      virtualMachine,
		ResultsChan:         make(chan verification.SectionResult),
		verification:        currentVerification,
	}
}

func (stageRunner *StageRunner) RunStages(sections []section.Section) error {
	for sectionNumber, section := range sections {
		shouldContinue, err := stageRunner.runSection(uint64(sectionNumber), section)
		if err != nil || !shouldContinue {
			return err
		}
	}
	return nil
}

func (stageRunner *StageRunner) runSection(sectionNumber uint64, section section.Section) (bool, error) {
	factorySuccess, err := stageRunner.runFactoryCommands(sectionNumber, section)
	if err != nil {
		return false, err
	}
	if !factorySuccess && !section.ContinueOnFailure() {
		return false, nil
	}
	commandsSuccess, err := stageRunner.runCommands(!factorySuccess, sectionNumber, section)

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

func (stageRunner *StageRunner) runFactoryCommands(sectionNumber uint64, sectionToRun section.Section) (bool, error) {
	var err error
	var command verification.Command
	sectionFailed := false

	factoryCommands := sectionToRun.FactoryCommands(false)

	for command, err = factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
		if stageRunner.verification.VerificationStatus == "cancelled" {
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

		stageRunId, err := stageRunner.resourcesConnection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return false, err
		}

		outputBuffer := new(bytes.Buffer)
		syncOutputBuffer := &syncWriter{writer: outputBuffer}

		consoleWriter := newConsoleTextWriter(stageRunner.resourcesConnection.Stages.Update, stageRunId)
		syncConsoleWriter := &syncWriter{writer: consoleWriter}

		returnCode, runErr := stageRunner.runCommand(command, os.Stdin, io.MultiWriter(syncOutputBuffer, syncConsoleWriter, os.Stdout), io.MultiWriter(syncOutputBuffer, syncConsoleWriter, os.Stderr))
		factoryCommands.Done()
		closeErr := consoleWriter.Close()

		stageFailed := returnCode != 0 || runErr != nil
		if stageFailed && !sectionFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}

		stageRunner.resourcesConnection.Stages.Update.SetStartTime(stageRunId, time.Now())
		setEndTimeErr := stageRunner.resourcesConnection.Stages.Update.SetEndTime(stageRunId, time.Now())
		setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRunId, returnCode)

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

func (stageRunner *StageRunner) runCommands(sectionPreviouslyFailed bool, sectionNumber uint64, sectionToRun section.Section) (bool, error) {
	var err error
	var command verification.Command
	sectionFailed := false

	commands := sectionToRun.Commands(false)
	for command, err = commands.Next(); err == nil; command, err = commands.Next() {
		if stageRunner.verification.VerificationStatus == "cancelled" {
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

		stageRunId, err := stageRunner.resourcesConnection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return false, err
		}

		consoleWriter := newConsoleTextWriter(stageRunner.resourcesConnection.Stages.Update, stageRunId)
		syncConsoleWriter := &syncWriter{writer: consoleWriter}

		returnCode, runErr := stageRunner.runCommand(command, os.Stdin, io.MultiWriter(syncConsoleWriter, os.Stdout), io.MultiWriter(syncConsoleWriter, os.Stderr))
		commands.Done()
		closeErr := consoleWriter.Close()

		stageFailed := returnCode != 0 || runErr != nil
		if stageFailed && !sectionFailed && !sectionPreviouslyFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}

		stageRunner.resourcesConnection.Stages.Update.SetStartTime(stageRunId, time.Now())
		setEndTimeErr := stageRunner.resourcesConnection.Stages.Update.SetEndTime(stageRunId, time.Now())
		setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRunId, returnCode)

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
	locker sync.Mutex
}

func (writer *syncWriter) Write(bytes []byte) (int, error) {
	writer.locker.Lock()
	defer writer.locker.Unlock()
	return writer.writer.Write(bytes)
}

type consoleTextWriter struct {
	stagesUpdateHandler resources.StagesUpdateHandler
	stageRunId          uint64
	buffer              bytes.Buffer
	locker              sync.Mutex
	closeChan           chan bool
	lastLine            string
	lastLineNumber      uint64
}

func newConsoleTextWriter(stagesUpdateHandler resources.StagesUpdateHandler, stageRunId uint64) *consoleTextWriter {
	var buffer bytes.Buffer
	var locker sync.Mutex
	writer := &consoleTextWriter{stagesUpdateHandler, stageRunId, buffer, locker, make(chan bool, 1), "", 1}
	go writer.flushOnTick()
	return writer
}

func (writer *consoleTextWriter) flushOnTick() {
	ticker := time.NewTicker(250 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			writer.flush()
		case <-writer.closeChan:
			ticker.Stop()
			writer.flush()
			return
		}
	}
}

func (writer *consoleTextWriter) Write(bytes []byte) (int, error) {
	writer.locker.Lock()
	defer writer.locker.Unlock()

	numBytes, err := writer.buffer.Write(bytes)

	return numBytes, err
}

func (writer *consoleTextWriter) Close() error {
	writer.closeChan <- true
	close(writer.closeChan)
	return nil
}

func (writer *consoleTextWriter) flush() error {
	writer.locker.Lock()
	defer writer.locker.Unlock()

	if writer.buffer.Len() == 0 {
		return nil
	}

	lines := strings.Split(writer.buffer.String(), "\n")
	linesMap := make(map[uint64]string, len(lines))

	firstLineEmpty := lines[0] == ""

	lines[0] = writer.lastLine + lines[0]

	for index, line := range lines {
		linesMap[writer.lastLineNumber+uint64(index)] = line
	}

	if firstLineEmpty {
		delete(linesMap, writer.lastLineNumber)
	}

	if strings.TrimSpace(lines[len(lines)-1]) == "" {
		delete(linesMap, writer.lastLineNumber+uint64(len(lines)-1))
	}

	writer.buffer.Reset()

	writer.lastLine = lines[len(lines)-1]
	writer.lastLineNumber = writer.lastLineNumber + uint64(len(lines)) - 1

	err := writer.stagesUpdateHandler.AddConsoleLines(writer.stageRunId, linesMap)
	return err
}
