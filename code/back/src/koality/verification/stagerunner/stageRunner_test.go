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

func (s *StageRunnerSuite) SetUpTest(c *gocheck.C) {
	err := database.PopulateDatabase()
	c.Assert(err, gocheck.IsNil)

	s.resourcesConnection, err = database.New()
	c.Assert(err, gocheck.IsNil)

	s.virtualMachine = localmachine.New()
	s.repository, err = s.resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	c.Assert(err, gocheck.IsNil)

	s.verification, err = s.resourcesConnection.Verifications.Create.Create(s.repository.Id, "1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	c.Assert(err, gocheck.IsNil)

	s.stageRunner = New(s.resourcesConnection, s.virtualMachine, s.verification)
}

func (s *StageRunnerSuite) TearDownTest(c *gocheck.C) {
	s.virtualMachine.Terminate()
}

func (s *StageRunnerSuite) TestSimplePassingStages(c *gocheck.C) {
	var err error
	commands := []verification.Command{verification.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err = s.resourcesConnection.Stages.Create.Create(s.verification.Id, 0, command.Name(), uint64(index))
		c.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range s.stageRunner.ResultsChan {
			c.Check(result.Passed, gocheck.Equals, true,
				fmt.Sprintf("Failed section %s", result.Section))
		}

		doneChan <- err
	}(doneChan)

	err = s.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	c.Assert(err, gocheck.IsNil)

	close(s.stageRunner.ResultsChan)
}

func (s *StageRunnerSuite) TestExporting(c *gocheck.C) {
	err := database.PopulateDatabase()
	c.Assert(err, gocheck.IsNil)

	resourcesConnection, err := database.New()
	c.Assert(err, gocheck.IsNil)

	virtualMachine := localmachine.New()
	defer virtualMachine.Terminate()

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	c.Assert(err, gocheck.IsNil)

	// REVIEW(dhuang) is putting this here a good idea?
	_, err = resourcesConnection.Settings.Update.SetS3ExporterSettings(
		"AKIAJIXWHV32ZY75SQBQ", "JgD4KK376m9Z3E3MjMt8YcPg3cuzl958Qjtbrht1", "koality-whim")
	c.Assert(err, gocheck.IsNil)

	copyExec, err := virtualMachine.FileCopy("$HOME/code/back/src/koality/util/xunitsamples/*.xml", "xunitPath/")
	c.Assert(err, gocheck.IsNil)

	err = copyExec.Run()
	c.Assert(err, gocheck.IsNil)

	compileCmd := exec.Command("go", "build", "-o",
		"/home/koality/code/back/src/koality/util/exportPaths",
		"/home/koality/code/back/src/koality/util/exportPaths.go")
	err = compileCmd.Run()
	c.Assert(err, gocheck.IsNil)

	commands := []verification.Command{verification.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err = s.resourcesConnection.Stages.Create.Create(s.verification.Id, 0, command.Name(), uint64(index))
		c.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)
	usr, err := user.Current()
	c.Assert(err, gocheck.IsNil)

	exportPaths := []string{
		path.Join(usr.HomeDir, "code", "back", "src", "koality", "util", "xunitsamples"),
	}

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, exportPaths)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range s.stageRunner.ResultsChan {
			c.Check(result.Passed, gocheck.Equals, true,
				fmt.Sprintf("Failed section %s", result.Section))
		}

		doneChan <- err
	}(doneChan)

	err = s.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	c.Assert(err, gocheck.IsNil)

	stage, err := s.resourcesConnection.Stages.Read.GetBySectionNumberAndName(s.verification.Id, 0, "pass")
	c.Assert(err, gocheck.IsNil)

	stageRuns, err := s.resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	c.Assert(err, gocheck.IsNil)

	c.Assert(stageRuns, gocheck.HasLen, 1)

	exports, err := s.resourcesConnection.Stages.Read.GetExports(stageRuns[0].Id)
	c.Assert(err, gocheck.IsNil)

	c.Assert(exports, gocheck.HasLen, 1)
	expectedBucket := "koality-whim"
	expectedPath := "/home/koality/code/back/src/koality/util/xunitsamples/sample1.xml"
	expectedKey :=
		fmt.Sprintf("repository/%d/verification/%d/stage/%d/stageRun/%d%s",
			s.repository.Id, s.verification.Id, stage.Id,
			stageRuns[0].Id, expectedPath)
	expectedExport := resources.Export{BucketName: expectedBucket, Path: expectedPath, Key: expectedKey}
	c.Assert(exports[0], gocheck.Equals, expectedExport)

	close(s.stageRunner.ResultsChan)
}

func (s *StageRunnerSuite) TestXunitParser(c *gocheck.C) {
	copyExec, err := s.virtualMachine.FileCopy("$HOME/code/back/src/koality/util/xunitsamples/*.xml", "xunitPath/")
	c.Assert(err, gocheck.IsNil)

	err = copyExec.Run()
	c.Assert(err, gocheck.IsNil)

	compileCmd := exec.Command("go", "build", "-o",
		"/home/koality/code/back/src/koality/util/getXunitResults",
		"/home/koality/code/back/src/koality/util/getXunitResults.go")
	err = compileCmd.Run()
	c.Assert(err, gocheck.IsNil)

	commands := []verification.Command{
		remotecommand.NewRemoteCommand(false, "command name!!", 1000, []string{"xunitPath"}, []string{"true"}),
	}
	for index, command := range commands {
		_, err := s.resourcesConnection.Stages.Create.Create(s.verification.Id, 0, command.Name(), uint64(index))
		c.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range s.stageRunner.ResultsChan {
			c.Check(result.Passed, gocheck.Equals, true,
				fmt.Sprintf("Failed section %s", result.Section))
		}

		doneChan <- err
	}(doneChan)

	err = s.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	c.Assert(err, gocheck.IsNil)

	stage, err := s.resourcesConnection.Stages.Read.GetBySectionNumberAndName(s.verification.Id, 0, "command name!!")
	c.Assert(err, gocheck.IsNil)

	stageRuns, err := s.resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stageRuns, gocheck.HasLen, 1)

	xunitResults, err := s.resourcesConnection.Stages.Read.GetAllXunitResults(stageRuns[0].Id)
	c.Assert(err, gocheck.IsNil)

	usr, err := user.Current()
	c.Assert(err, gocheck.IsNil)

	xunitPath := path.Join(usr.HomeDir, "code", "back", "src", "koality", "util", "xunitsamples")
	expectedXunitBytes, err := xunit.GetXunitResults("*.xml", []string{xunitPath}, ioutil.Discard, os.Stderr)
	c.Assert(err, gocheck.IsNil)

	var expectedXunitResults []resources.XunitResult
	err = json.Unmarshal(expectedXunitBytes, &expectedXunitResults)
	c.Assert(err, gocheck.IsNil)

	// normalize paths
	for i := range xunitResults {
		xunitResults[i].Path = filepath.Base(xunitResults[i].Path)
	}
	for i := range expectedXunitResults {
		expectedXunitResults[i].Path = filepath.Base(expectedXunitResults[i].Path)
	}
	c.Assert(xunitResults, gocheck.Equals, expectedXunitResults)

	close(s.stageRunner.ResultsChan)
}

func (s *StageRunnerSuite) TestSimpleFailingStages(c *gocheck.C) {
	commands := []verification.Command{verification.NewShellCommand("fail", shell.Command("false"))}
	for index, command := range commands {
		_, err := s.resourcesConnection.Stages.Create.Create(s.verification.Id, 0, command.Name(), uint64(index))
		c.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range s.stageRunner.ResultsChan {
			c.Check(result.Passed, gocheck.Equals, false,
				fmt.Sprintf("Passed section %s", result.Section))
		}

		doneChan <- nil
	}(doneChan)

	err := s.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	c.Assert(err, gocheck.IsNil)

	close(s.stageRunner.ResultsChan)
}
