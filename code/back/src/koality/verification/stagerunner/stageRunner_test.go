package stagerunner

import (
	"encoding/json"
	"fmt"
	"github.com/ashokgelal/gocheck"
	"io/ioutil"
	"koality/resources"
	"koality/resources/database"
	"koality/shell"
	"koality/util/xunit"
	"koality/verification"
	"koality/verification/config/commandgroup"
	"koality/verification/config/remotecommand"
	"koality/verification/config/section"
	"koality/vm"
	"koality/vm/localmachine"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type StageRunnerSuite struct {
	resourcesConnection *resources.Connection
	virtualMachine      vm.VirtualMachine
	repository          *resources.Repository
	verification        *resources.Verification
	stageRunner         *StageRunner
}

var _ = gocheck.Suite(&StageRunnerSuite{})

func (suite *StageRunnerSuite) SetUpTest(check *gocheck.C) {
	err := database.PopulateDatabase()
	check.Assert(err, gocheck.IsNil)

	suite.resourcesConnection, err = database.New()
	check.Assert(err, gocheck.IsNil)

	suite.virtualMachine = localmachine.New()
	suite.repository, err = suite.resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	check.Assert(err, gocheck.IsNil)

	suite.verification, err = suite.resourcesConnection.Verifications.Create.Create(suite.repository.Id,
		"1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	check.Assert(err, gocheck.IsNil)

	suite.stageRunner = New(suite.resourcesConnection, suite.virtualMachine, suite.verification)
}

func (suite *StageRunnerSuite) TearDownTest(check *gocheck.C) {
	if suite.virtualMachine != nil {
		suite.virtualMachine.Terminate()
	}
}

func (suite *StageRunnerSuite) TestSimplePassingStages(check *gocheck.C) {
	var err error
	commands := []verification.Command{verification.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err = suite.resourcesConnection.Stages.Create.Create(suite.verification.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			check.Check(result.Passed, gocheck.Equals, true,
				fmt.Sprintf("Failed section %s", result.Section))
		}

		doneChan <- err
	}(doneChan)

	err = suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	close(suite.stageRunner.ResultsChan)
}

func (suite *StageRunnerSuite) TestExporting(check *gocheck.C) {
	// REVIEW(dhuang) is putting this here a good idea?
	_, err := suite.resourcesConnection.Settings.Update.SetS3ExporterSettings(
		"AKIAJIXWHV32ZY75SQBQ", "JgD4KK376m9Z3E3MjMt8YcPg3cuzl958Qjtbrht1", "koality-whim")
	check.Assert(err, gocheck.IsNil)

	copyExec, err := suite.virtualMachine.FileCopy("$HOME/code/back/src/koality/util/xunitsamples/*.xml", "xunitPath/")
	check.Assert(err, gocheck.IsNil)

	err = copyExec.Run()
	check.Assert(err, gocheck.IsNil)

	compileCmd := exec.Command("go", "build", "-o",
		"/home/koality/code/back/src/koality/util/exportPaths",
		"/home/koality/code/back/src/koality/util/exportPaths.go")
	err = compileCmd.Run()
	check.Assert(err, gocheck.IsNil)

	commands := []verification.Command{verification.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err = suite.resourcesConnection.Stages.Create.Create(suite.verification.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)
	usr, err := user.Current()
	check.Assert(err, gocheck.IsNil)

	exportPaths := []string{
		path.Join(usr.HomeDir, "code", "back", "src", "koality", "util", "xunitsamples"),
	}

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, exportPaths)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			// check.Check(result.Passed, gocheck.Equals, true, gocheck.Commentf("Failed section %s", result.Section))
			check.Check(result.Passed, gocheck.Equals, true)
		}

		doneChan <- err
	}(doneChan)

	err = suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	stage, err := suite.resourcesConnection.Stages.Read.GetBySectionNumberAndName(suite.verification.Id, 0, "pass")
	check.Assert(err, gocheck.IsNil)

	stageRuns, err := suite.resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	check.Assert(err, gocheck.IsNil)

	// check.Assert(stageRuns, gocheck.HasLen, 1)
	check.Assert(len(stageRuns), gocheck.Equals, 1)

	exports, err := suite.resourcesConnection.Stages.Read.GetExports(stageRuns[0].Id)
	check.Assert(err, gocheck.IsNil)

	// check.Assert(exports, gocheck.HasLen, 1)
	check.Assert(len(exports), gocheck.Equals, 1)
	expectedBucket := "koality-whim"
	expectedPath := "/home/koality/code/back/src/koality/util/xunitsamples/sample1.xml"
	expectedKey :=
		fmt.Sprintf("repository/%d/verification/%d/stage/%d/stageRun/%d%s",
			suite.repository.Id, suite.verification.Id, stage.Id,
			stageRuns[0].Id, expectedPath)
	expectedExport := resources.Export{BucketName: expectedBucket, Path: expectedPath, Key: expectedKey}
	check.Assert(exports[0], gocheck.Equals, expectedExport)

	close(suite.stageRunner.ResultsChan)
}

func (suite *StageRunnerSuite) TestXunitParser(check *gocheck.C) {
	copyExec, err := suite.virtualMachine.FileCopy("$HOME/code/back/src/koality/util/xunitsamples/*.xml", "xunitPath/")
	check.Assert(err, gocheck.IsNil)

	err = copyExec.Run()
	check.Assert(err, gocheck.IsNil)

	compileCmd := exec.Command("go", "build", "-o",
		"/home/koality/code/back/src/koality/util/getXunitResults",
		"/home/koality/code/back/src/koality/util/getXunitResults.go")
	err = compileCmd.Run()
	check.Assert(err, gocheck.IsNil)

	commands := []verification.Command{
		remotecommand.NewRemoteCommand(false, "command name!!", 1000, []string{"xunitPath"}, []string{"true"}),
	}
	for index, command := range commands {
		_, err := suite.resourcesConnection.Stages.Create.Create(suite.verification.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			// check.Check(result.Passed, gocheck.Equals, true, gocheck.Commentf("Failed section %s", result.Section))
			check.Check(result.Passed, gocheck.Equals, true)
		}

		doneChan <- err
	}(doneChan)

	err = suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	stage, err := suite.resourcesConnection.Stages.Read.GetBySectionNumberAndName(suite.verification.Id, 0, "command name!!")
	check.Assert(err, gocheck.IsNil)

	stageRuns, err := suite.resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	check.Assert(err, gocheck.IsNil)
	// check.Assert(stageRuns, gocheck.HasLen, 1)
	check.Assert(len(stageRuns), gocheck.Equals, 1)

	xunitResults, err := suite.resourcesConnection.Stages.Read.GetAllXunitResults(stageRuns[0].Id)
	check.Assert(err, gocheck.IsNil)

	usr, err := user.Current()
	check.Assert(err, gocheck.IsNil)

	xunitPath := path.Join(usr.HomeDir, "code", "back", "src", "koality", "util", "xunitsamples")
	expectedXunitBytes, err := xunit.GetXunitResults("*.xml", []string{xunitPath}, ioutil.Discard, os.Stderr)
	check.Assert(err, gocheck.IsNil)

	var expectedXunitResults []resources.XunitResult
	err = json.Unmarshal(expectedXunitBytes, &expectedXunitResults)
	check.Assert(err, gocheck.IsNil)

	// normalize paths
	for i := range xunitResults {
		xunitResults[i].Path = filepath.Base(xunitResults[i].Path)
	}
	for i := range expectedXunitResults {
		expectedXunitResults[i].Path = filepath.Base(expectedXunitResults[i].Path)
	}
	check.Assert(xunitResults, gocheck.Equals, expectedXunitResults)

	close(suite.stageRunner.ResultsChan)
}

func (suite *StageRunnerSuite) TestSimpleFailingStages(check *gocheck.C) {
	commands := []verification.Command{verification.NewShellCommand("fail", shell.Command("false"))}
	for index, command := range commands {
		_, err := suite.resourcesConnection.Stages.Create.Create(suite.verification.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			// check.Check(result.Passed, gocheck.Equals, false, gocheck.Commentf("Passed section %s", result.Section))
			check.Check(result.Passed, gocheck.Equals, false)
		}

		doneChan <- nil
	}(doneChan)

	err := suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	close(suite.stageRunner.ResultsChan)
}
