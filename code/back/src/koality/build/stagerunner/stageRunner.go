package stagerunner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/LessThanThreeLabs/go.crypto/ssh"
	"github.com/dchest/goyaml"
	"io"
	"koality/build"
	"koality/build/config"
	"koality/build/config/commandgroup"
	"koality/build/config/section"
	"koality/resources"
	"koality/shell"
	"koality/shell/shellutil"
	"koality/util/pathtranslator"
	"koality/vm"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type StageRunner struct {
	ResultsChan         chan build.SectionResult
	resourcesConnection *resources.Connection
	virtualMachine      vm.VirtualMachine
	build               *resources.Build
	exporter            Exporter
}

func New(resourcesConnection *resources.Connection, virtualMachine vm.VirtualMachine, currentBuild *resources.Build, exporter Exporter) *StageRunner {
	// exporter.
	return &StageRunner{
		ResultsChan:         make(chan build.SectionResult),
		resourcesConnection: resourcesConnection,
		virtualMachine:      virtualMachine,
		build:               currentBuild,
		exporter:            exporter,
	}
}

func (stageRunner *StageRunner) RunStages(sections, finalSections []section.Section, environment map[string]string) error {
	for sectionNumber, section := range sections {
		shouldContinue, err := stageRunner.runSection(uint64(sectionNumber), section, environment)
		if err != nil {
			stageRunner.ResultsChan <- build.SectionResult{
				Section:       section.Name(),
				Final:         section.IsFinal(),
				FailSectionOn: section.FailOn(),
				Passed:        false,
			}
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
	for stageRunner.build.Status == "running" {
		time.Sleep(time.Second)
	}
	if stageRunner.build.Status == "cancelled" {
		return nil
	}
	if environment == nil {
		environment = make(map[string]string)
	}
	environment["KOALITY_STATUS"] = stageRunner.build.Status
	for finalSectionNumber, finalSection := range finalSections {
		// Leave room for potential merge stage
		sectionNumber := uint64(finalSectionNumber + len(sections) + 1)
		shouldContinue, err := stageRunner.runSection(sectionNumber, finalSection, environment)
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

	exportSucess, err := stageRunner.runExports(sectionNumber, section, environment)
	if err != nil {
		return false, err
	}

	if factorySuccess && commandsSuccess && exportSucess {
		stageRunner.ResultsChan <- build.SectionResult{
			Section:       section.Name(),
			Final:         section.IsFinal(),
			FailSectionOn: section.FailOn(),
			Passed:        true,
		}
	}
	return commandsSuccess || section.ContinueOnFailure(), nil
}

func (stageRunner *StageRunner) copyAndRunExecOnVm(stageRunId uint64, execName string, args []string, environment map[string]string) ([]byte, error) {
	binaryPath, err := pathtranslator.TranslatePathWithCheckFunc(pathtranslator.BinaryPath(execName), pathtranslator.CheckExecutable)
	if err != nil {
		return nil, err
	}

	copyExecutable, err := shellutil.CreateScpExecutable(binaryPath, execName)
	if err != nil {
		return nil, err
	}

	copyExecution, err := stageRunner.virtualMachine.Execute(copyExecutable)
	if err != nil {
		return nil, err
	}

	if err = copyExecution.Wait(); err != nil {
		return nil, err
	}

	cmd := shell.Commandf("./%s %s", execName, strings.Join(args, " "))
	writeBuffer := new(bytes.Buffer)
	consoleTextWriter, err := continuedConsoleTextWriter(stageRunner.resourcesConnection.Stages.Read, stageRunner.resourcesConnection.Stages.Update, stageRunId)
	if err != nil {
		return nil, err
	}

	runExecutable := shell.Executable{
		Command:     cmd,
		Stdout:      writeBuffer,
		Stderr:      io.MultiWriter(consoleTextWriter, os.Stderr),
		Environment: environment,
	}

	runExecution, err := stageRunner.virtualMachine.Execute(runExecutable)
	if err != nil {
		return nil, err
	}

	if err = runExecution.Wait(); err != nil {
		return nil, err
	}

	return writeBuffer.Bytes(), nil
}

func (stageRunner *StageRunner) getXunitResults(stageRunId uint64, command build.Command, environment map[string]string) ([]resources.XunitResult, error) {
	directories := command.XunitPaths()
	if len(directories) == 0 {
		return nil, nil
	}
	args := []string{shell.Quote("*.xml"), strings.Join(directories, " ")}
	output, err := stageRunner.copyAndRunExecOnVm(stageRunId, "getXunitResults", args, environment)
	if err != nil {
		return nil, err
	}

	var res []resources.XunitResult
	if err = json.Unmarshal(output, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (stageRunner *StageRunner) runFactoryCommands(sectionNumber uint64, sectionToRun section.Section, environment map[string]string) (bool, error) {
	var err error
	var command build.Command
	sectionFailed := false

	factoryCommands := sectionToRun.FactoryCommands(false)

	for command, err = factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
		if stageRunner.build.Status == "cancelled" {
			return false, nil
		}

		stages, err := stageRunner.resourcesConnection.Stages.Read.GetAll(stageRunner.build.Id)
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
			return false, fmt.Errorf("Unable to find a stage to match %#v", command)
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
			stageRunner.ResultsChan <- build.SectionResult{
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

		repository, err := stageRunner.resourcesConnection.Repositories.Read.Get(stageRunner.build.RepositoryId)
		if err != nil {
			// TODO (bbland): this is a fatal error.
			return false, err
		}

		newCommands, err := config.ParseRemoteCommands(yamlParsedCommands, true, repository.Name)
		if err != nil {
			// TODO (bbland): display the error to the user
			return false, err
		}
		stages, err = stageRunner.resourcesConnection.Stages.Read.GetAll(stageRunner.build.Id)
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

			_, err = stageRunner.resourcesConnection.Stages.Create.Create(stageRunner.build.Id, sectionNumber, command.Name(), orderNumber)
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
	var command build.Command
	sectionFailed := false

	commands := sectionToRun.Commands(false)
	for command, err = commands.Next(); err == nil; command, err = commands.Next() {
		if stageRunner.build.Status == "cancelled" {
			return false, nil
		}

		stages, err := stageRunner.resourcesConnection.Stages.Read.GetAll(stageRunner.build.Id)
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
			return false, fmt.Errorf("Unable to find a stage to match %#v", command)
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
			stageRunner.ResultsChan <- build.SectionResult{
				Section:       sectionToRun.Name(),
				Final:         sectionToRun.IsFinal(),
				FailSectionOn: sectionToRun.FailOn(),
				Passed:        false,
			}
		}

		setEndTimeErr := stageRunner.resourcesConnection.Stages.Update.SetEndTime(stageRun.Id, time.Now())
		setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRun.Id, returnCode)
		xunitResults, xunitErr := stageRunner.getXunitResults(stageRun.Id, command, environment)
		if xunitErr != nil {
			return false, xunitErr
		}

		xunitErr = stageRunner.resourcesConnection.Stages.Update.AddXunitResults(stageRun.Id, xunitResults)
		if xunitErr != nil {
			return false, xunitErr
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

func (stageRunner *StageRunner) runExports(sectionNumber uint64, sectionToRun section.Section, environment map[string]string) (bool, error) {
	exportPaths := sectionToRun.Exports()
	if len(exportPaths) == 0 {
		return true, nil
	}
	if stageRunner.build.Status == "cancelled" {
		return false, nil
	}

	stages, err := stageRunner.resourcesConnection.Stages.Read.GetAll(stageRunner.build.Id)
	if err != nil {
		return false, err
	}

	var stageId uint64

	for _, stage := range stages {
		if stage.SectionNumber == sectionNumber && stage.Name == sectionToRun.Name()+".export" {
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
	err = stageRunner.resourcesConnection.Stages.Update.SetStartTime(stageRun.Id, time.Now())
	if err != nil {
		return false, err
	}

	exports, exportErr := stageRunner.exporter.ExportAndGetResults(stageId, stageRun.Id, stageRunner, exportPaths, environment)
	if exportErr != nil {
		return false, exportErr
	}

	setEndTimeErr := stageRunner.resourcesConnection.Stages.Update.SetEndTime(stageRun.Id, time.Now())
	setReturnCodeErr := stageRunner.resourcesConnection.Stages.Update.SetReturnCode(stageRun.Id, 0)
	exportErr = stageRunner.resourcesConnection.Stages.Update.AddExports(stageRun.Id, exports)

	if setEndTimeErr != nil {
		return false, setEndTimeErr
	}
	if setReturnCodeErr != nil {
		return false, setReturnCodeErr
	}
	if exportErr != nil {
		return false, setEndTimeErr
	}
	return true, nil
}

func (stageRunner *StageRunner) runCommand(command build.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (int, error) {
	executable := command.Executable()

	if stdin != nil {
		if executable.Stdin != nil {
			executable.Stdin = io.MultiReader(executable.Stdin, stdin)
		} else {
			executable.Stdin = stdin
		}
	}
	if stdout != nil {
		if executable.Stdout != nil {
			executable.Stdout = io.MultiWriter(executable.Stdout, stdout)
		} else {
			executable.Stdout = stdout
		}
	}
	if stderr != nil {
		if executable.Stderr != nil {
			executable.Stderr = io.MultiWriter(executable.Stderr, stderr)
		} else {
			executable.Stderr = stderr
		}
	}
	if environment != nil {
		if executable.Environment == nil {
			executable.Environment = environment
		} else {
			for key, value := range environment {
				executable.Environment[key] = value
			}
		}
	}

	execution, err := stageRunner.virtualMachine.Execute(executable)
	if err != nil {
		return 255, err
	}

	exitErr := execution.Wait()
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
