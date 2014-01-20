package stagerunner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"koality/resources"
	"koality/resources/database"
	"koality/shell"
	"koality/util/xunit"
	"koality/verification"
	"koality/verification/config/commandgroup"
	"koality/verification/config/remotecommand"
	"koality/verification/config/section"
	"koality/vm/localmachine"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"testing"
)

func TestSimplePassingStages(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}

	virtualMachine := localmachine.New()
	defer virtualMachine.Terminate()

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	if err != nil {
		test.Fatal(err)
	}

	currentVerification, err := resourcesConnection.Verifications.Create.Create(repository.Id, "1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	stageRunner := New(resourcesConnection, virtualMachine, currentVerification)

	commands := []verification.Command{verification.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err := resourcesConnection.Stages.Create.Create(currentVerification.Id, 0, command.Name(), uint64(index))
		if err != nil {
			test.Fatal(err)
		}
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range stageRunner.ResultsChan {
			if !result.Passed {
				test.Log(fmt.Sprintf("Failed section %s", result.Section))
				test.Fail()
			}
		}

		doneChan <- err
	}(doneChan)

	err = stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	if err != nil {
		test.Fatal(err)
	}

	close(stageRunner.ResultsChan)
}

func TestExporting(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}

	virtualMachine := localmachine.New()
	defer virtualMachine.Terminate()

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	if err != nil {
		test.Fatal(err)
	}

	// REVIEW(dhuang) is putting this here a good idea?
	_, err = resourcesConnection.Settings.Update.SetS3ExporterSettings(
		"AKIAJIXWHV32ZY75SQBQ", "JgD4KK376m9Z3E3MjMt8YcPg3cuzl958Qjtbrht1", "koality-whim")
	if err != nil {
		test.Fatal(err)
	}

	copyExec, err := virtualMachine.FileCopy("$HOME/code/back/src/koality/util/xunitsamples/*.xml", "xunitPath/")
	if err != nil {
		test.Fatal(err)
	}

	if err = copyExec.Run(); err != nil {
		test.Fatal(err)
	}

	currentVerification, err := resourcesConnection.Verifications.Create.Create(repository.Id, "1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	compileCmd := exec.Command("go", "build", "-o",
		"/home/koality/code/back/src/koality/util/exportPaths",
		"/home/koality/code/back/src/koality/util/exportPaths.go")
	if err = compileCmd.Run(); err != nil {
		test.Fatal(err)
	}

	stageRunner := New(resourcesConnection, virtualMachine, currentVerification)

	commands := []verification.Command{verification.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err := resourcesConnection.Stages.Create.Create(currentVerification.Id, 0, command.Name(), uint64(index))
		if err != nil {
			test.Fatal(err)
		}
	}

	commandGroup := commandgroup.New(commands)
	usr, err := user.Current()
	if err != nil {
		test.Fatal(err)
	}
	exportPaths := []string{
		path.Join(usr.HomeDir, "code", "back", "src", "koality", "util", "xunitsamples"),
	}

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, exportPaths)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range stageRunner.ResultsChan {
			if !result.Passed {
				test.Log(fmt.Sprintf("Failed section %s", result.Section))
				test.Fail()
			}
		}

		doneChan <- err
	}(doneChan)

	err = stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	if err != nil {
		test.Fatal(err)
	}

	stage, err := resourcesConnection.Stages.Read.GetBySectionNumberAndName(currentVerification.Id, 0, "pass")
	stageRuns, err := resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	if len(stageRuns) != 1 {
		test.Fatal("expected there to be exactly one stage run")
	}

	exports, err := resourcesConnection.Stages.Read.GetExports(stageRuns[0].Id)
	if err != nil {
		test.Fatal(err)
	}

	if len(exports) != 1 {
		test.Fatal("expected 1 export, got", len(exports))
	}
	expectedBucket := "koality-whim"
	expectedPath := "/home/koality/code/back/src/koality/util/xunitsamples/sample1.xml"
	expectedKey :=
		fmt.Sprintf("repository/%d/verification/%d/stage/%d/stageRun/%d%s",
			repository.Id, currentVerification.Id, stage.Id,
			stageRuns[0].Id, expectedPath)
	expectedExport := resources.Export{BucketName: expectedBucket, Path: expectedPath, Key: expectedKey}
	if exports[0] != expectedExport {
		test.Fatal("expected", expectedExport, "\nbut got", exports[0])
	}

	close(stageRunner.ResultsChan)
}

func TestXunitParser(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}

	virtualMachine := localmachine.New()
	defer virtualMachine.Terminate()

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	if err != nil {
		test.Fatal(err)
	}

	copyExec, err := virtualMachine.FileCopy("$HOME/code/back/src/koality/util/xunitsamples/*.xml", "xunitPath/")
	if err != nil {
		test.Fatal(err)
	}

	if err = copyExec.Run(); err != nil {
		test.Fatal(err)
	}

	currentVerification, err := resourcesConnection.Verifications.Create.Create(repository.Id, "1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	compileCmd := exec.Command("go", "build", "-o",
		"/home/koality/code/back/src/koality/util/getXunitResults",
		"/home/koality/code/back/src/koality/util/getXunitResults.go")
	if err = compileCmd.Run(); err != nil {
		test.Fatal(err)
	}

	stageRunner := New(resourcesConnection, virtualMachine, currentVerification)

	commands := []verification.Command{
		remotecommand.NewRemoteCommand(false, "command name!!", 1000, []string{"xunitPath"}, []string{"true"}),
	}
	for index, command := range commands {
		_, err := resourcesConnection.Stages.Create.Create(currentVerification.Id, 0, command.Name(), uint64(index))
		if err != nil {
			test.Fatal(err)
		}
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range stageRunner.ResultsChan {
			if !result.Passed {
				test.Log(fmt.Sprintf("Failed section %s", result.Section))
				test.Fail()
			}
		}

		doneChan <- err
	}(doneChan)

	err = stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	if err != nil {
		test.Fatal(err)
	}

	stage, err := resourcesConnection.Stages.Read.GetBySectionNumberAndName(currentVerification.Id, 0, "command name!!")
	if err != nil {
		test.Fatal(err)
	}

	stageRuns, err := resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	if len(stageRuns) != 1 {
		test.Fatal("expected there to be exactly one stage run")
	}

	xunitResults, err := resourcesConnection.Stages.Read.GetAllXunitResults(stageRuns[0].Id)
	if err != nil {
		test.Fatal(err)
	}

	usr, err := user.Current()
	if err != nil {
		test.Fatal(err)
	}

	xunitPath := path.Join(usr.HomeDir, "code", "back", "src", "koality", "util", "xunitsamples")
	expectedXunitBytes, err := xunit.GetXunitResults("*.xml", []string{xunitPath}, ioutil.Discard, os.Stderr)
	if err != nil {
		test.Fatal(err)
	}

	var expectedXunitResults []resources.XunitResult
	if err = json.Unmarshal(expectedXunitBytes, &expectedXunitResults); err != nil {
		test.Fatal(err)
	}

	// normalize paths
	for i := range xunitResults {
		xunitResults[i].Path = filepath.Base(xunitResults[i].Path)
	}
	for i := range expectedXunitResults {
		expectedXunitResults[i].Path = filepath.Base(expectedXunitResults[i].Path)
	}
	if len(xunitResults) != len(expectedXunitResults) || len(expectedXunitResults) != 2 {
		test.Fatal("different number of results than expected")
	}
	if xunitResults[0] != expectedXunitResults[0] || xunitResults[1] != expectedXunitResults[1] {
		test.Fatal("xunit results:\n", xunitResults, "did not match expected:\n", expectedXunitResults)
	}

	close(stageRunner.ResultsChan)
}

func TestSimpleFailingStages(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}

	virtualMachine := localmachine.New()
	defer virtualMachine.Terminate()

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	if err != nil {
		test.Fatal(err)
	}

	currentVerification, err := resourcesConnection.Verifications.Create.Create(repository.Id, "1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	stageRunner := New(resourcesConnection, virtualMachine, currentVerification)

	commands := []verification.Command{verification.NewShellCommand("fail", shell.Command("false"))}
	for index, command := range commands {
		_, err := resourcesConnection.Stages.Create.Create(currentVerification.Id, 0, command.Name(), uint64(index))
		if err != nil {
			test.Fatal(err)
		}
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range stageRunner.ResultsChan {
			if result.Passed {
				test.Log(fmt.Sprintf("Passed section %s", result.Section))
				test.Fail()
			}
		}

		doneChan <- err
	}(doneChan)

	err = stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	if err != nil {
		test.Fatal(err)
	}

	close(stageRunner.ResultsChan)
}
