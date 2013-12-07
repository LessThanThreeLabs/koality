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

	factoryCommands := sectionToRun.FactoryCommands()

	index := 0

	for command, err = factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
		if stageRunner.verification.VerificationStatus == "cancelled" {
			return false, nil
		}

		stageId, err := stageRunner.resourcesConnection.Stages.Create.Create(stageRunner.verification.Id, sectionNumber, command.Name(), uint64(index))
		if err != nil {
			return false, err
		}

		stageRunId, err := stageRunner.resourcesConnection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return false, err
		}

		outputBuffer := new(bytes.Buffer)
		syncOutputBuffer := &syncWriter{writer: outputBuffer}
		syncConsoleWriter := &syncWriter{writer: newConsoleTextWriter(stageRunner.resourcesConnection.Stages.Update, stageRunId)}
		returnCode, runErr := stageRunner.runCommand(command, os.Stdin, io.MultiWriter(syncOutputBuffer, syncConsoleWriter, os.Stdout), io.MultiWriter(syncOutputBuffer, syncConsoleWriter, os.Stderr))
		factoryCommands.Done()
		stageFailed := returnCode != 0 || runErr != nil
		if stageFailed && !sectionFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}

		setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRunId, returnCode)

		if runErr != nil {
			return false, runErr
		}
		if setReturnCodeErr != nil {
			return false, setReturnCodeErr
		}

		if stageFailed && !sectionToRun.ContinueOnFailure() {
			return false, nil
		}
		// TODO (bbland): parse the output into new commands
		newCommands := []verification.Command{}
		for _, newCommand := range newCommands {
			sectionToRun.AppendCommand(newCommand)
		}
		index++
	}
	if err != commandgroup.NoMoreCommands {
		return false, err
	}
	return !sectionFailed, nil
}

func (stageRunner *StageRunner) runCommands(sectionPreviouslyFailed bool, sectionNumber uint64, sectionToRun section.Section) (bool, error) {
	var err error
	var command verification.Command
	sectionFailed := false

	index := 0

	commands := sectionToRun.Commands()
	for command, err = commands.Next(); err == nil; command, err = commands.Next() {
		if stageRunner.verification.VerificationStatus == "cancelled" {
			return false, nil
		}

		stageId, err := stageRunner.resourcesConnection.Stages.Create.Create(stageRunner.verification.Id, sectionNumber, sectionToRun.Name(), uint64(index))
		if err != nil {
			return false, err
		}

		stageRunId, err := stageRunner.resourcesConnection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return false, err
		}

		syncConsoleWriter := &syncWriter{writer: newConsoleTextWriter(stageRunner.resourcesConnection.Stages.Update, stageRunId)}
		returnCode, runErr := stageRunner.runCommand(command, os.Stdin, io.MultiWriter(syncConsoleWriter, os.Stdout), io.MultiWriter(syncConsoleWriter, os.Stderr))
		commands.Done()
		stageFailed := returnCode != 0 || runErr != nil
		if stageFailed && !sectionFailed && !sectionPreviouslyFailed {
			sectionFailed = true
			stageRunner.ResultsChan <- verification.SectionResult{
				Section:       sectionToRun.Name(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}

		setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRunId, returnCode)

		if runErr != nil {
			return false, runErr
		}
		if setReturnCodeErr != nil {
			return false, setReturnCodeErr
		}

		if stageFailed && !sectionToRun.ContinueOnFailure() {
			return false, nil
		}
		index++
	}
	if err != commandgroup.NoMoreCommands {
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
	lastLine            string
	lastLineNumber      uint64
}

func newConsoleTextWriter(stagesUpdateHandler resources.StagesUpdateHandler, stageRunId uint64) *consoleTextWriter {
	var buffer bytes.Buffer
	var locker sync.Mutex
	writer := &consoleTextWriter{stagesUpdateHandler, stageRunId, buffer, locker, "", 1}
	go writer.flushOnTick()
	return writer
}

func (writer *consoleTextWriter) flushOnTick() {
	ticker := time.Tick(250 * time.Millisecond)
	for {
		<-ticker
		writer.flush()
	}
}

func (writer *consoleTextWriter) Write(bytes []byte) (int, error) {
	writer.locker.Lock()
	defer writer.locker.Unlock()

	numBytes, err := writer.buffer.Write(bytes)

	return numBytes, err
}

// Not sure if this is right
func (writer *consoleTextWriter) Close() error {
	return writer.flush()
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
