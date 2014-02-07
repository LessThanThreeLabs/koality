package stagerunner

import (
	"encoding/json"
	"github.com/LessThanThreeLabs/gocheck"
	"io/ioutil"
	"koality/build"
	"koality/build/config/commandgroup"
	"koality/build/config/remotecommand"
	"koality/build/config/section"
	"koality/resources"
	"koality/resources/database"
	"koality/shell"
	"koality/util/xunit"
	"koality/vm"
	"koality/vm/localmachine"
	"os"
	"os/exec"
	"path"
	"sort"
	"testing"
)

type exportsType []resources.Export

var (
	mockExports exportsType
	bucket      = "mybucket"
)

func (exports exportsType) Len() int {
	return len(exports)
}

func (exports exportsType) Less(i, j int) bool {
	return exports[i].BucketName+exports[i].Key < exports[j].BucketName+exports[j].Key
}

func (exports exportsType) Swap(i, j int) {
	exports[i], exports[j] = exports[j], exports[i]
}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type StageRunnerSuite struct {
	resourcesConnection *resources.Connection
	virtualMachine      vm.VirtualMachine
	repository          *resources.Repository
	build               *resources.Build
	stageRunner         *StageRunner
}

var _ = gocheck.Suite(&StageRunnerSuite{})

func (suite *StageRunnerSuite) SetUpTest(check *gocheck.C) {
	err := database.PopulateDatabase()
	check.Assert(err, gocheck.IsNil)

	suite.resourcesConnection, err = database.New()
	check.Assert(err, gocheck.IsNil)

	suite.virtualMachine, err = localmachine.New()
	check.Assert(err, gocheck.IsNil)

	suite.repository, err = suite.resourcesConnection.Repositories.Create.Create("repositoryName", "git", "remote@Uri")
	check.Assert(err, gocheck.IsNil)

	suite.build, err = suite.resourcesConnection.Builds.Create.Create(suite.repository.Id,
		"1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	check.Assert(err, gocheck.IsNil)

	mockExports = nil
	for _, path := range []string{"a/b/c", "a/bb", "a/b/d", "sdf"} {
		mockExports = append(mockExports, resources.Export{
			BucketName: bucket,
			Path:       path,
			Key:        "foo/bar/" + path,
		})
	}
}

func (suite *StageRunnerSuite) TearDownTest(check *gocheck.C) {
	suite.resourcesConnection.Close()

	if suite.virtualMachine != nil {
		suite.virtualMachine.Terminate()
	}
}

type MockNoExportsExporter struct{}

func (exporter *MockNoExportsExporter) ExportAndGetResults(stageId, stageRunId uint64, stageRunner *StageRunner, exportPaths []string, environment map[string]string) ([]resources.Export, error) {
	return nil, nil
}

type MockExporter struct{}

func (exporter *MockExporter) ExportAndGetResults(stageId, stageRunId uint64, stageRunner *StageRunner, exportPaths []string, environment map[string]string) ([]resources.Export, error) {
	return mockExports, nil
}

func (suite *StageRunnerSuite) TestSimplePassingStages(check *gocheck.C) {
	suite.stageRunner = New(suite.resourcesConnection, suite.virtualMachine,
		suite.build, new(MockNoExportsExporter))

	var err error
	commands := []build.Command{build.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err = suite.resourcesConnection.Stages.Create.Create(suite.build.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			check.Check(result.Passed, gocheck.Equals, true,
				gocheck.Commentf("Failed section %s", result.Section))
		}

		doneChan <- err
	}(doneChan)

	err = suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	close(suite.stageRunner.ResultsChan)
}

func (suite *StageRunnerSuite) TestExporting(check *gocheck.C) {
	suite.stageRunner = New(suite.resourcesConnection, suite.virtualMachine,
		suite.build, new(MockExporter))
	commands := []build.Command{build.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err := suite.resourcesConnection.Stages.Create.Create(suite.build.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}
	_, err := suite.resourcesConnection.Stages.Create.Create(suite.build.Id, 0, "test.export", uint64(len(commands)))
	check.Assert(err, gocheck.IsNil)

	commandGroup := commandgroup.New(commands)
	exportPaths := []string{"somePath"}
	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, exportPaths)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			check.Check(result.Passed, gocheck.Equals, true, gocheck.Commentf("Failed section %s", result.Section))
		}

		doneChan <- err
	}(doneChan)

	err = suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	stage, err := suite.resourcesConnection.Stages.Read.GetBySectionNumberAndName(suite.build.Id, 0, "test.export")
	check.Assert(err, gocheck.IsNil)

	stageRuns, err := suite.resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	check.Assert(err, gocheck.IsNil)
	check.Assert(stageRuns, gocheck.HasLen, 1)

	var exports exportsType
	exports, err = suite.resourcesConnection.Stages.Read.GetAllExports(stageRuns[0].Id)
	check.Assert(err, gocheck.IsNil)

	sort.Sort(exports)
	sort.Sort(mockExports)
	check.Assert(exports, gocheck.DeepEquals, mockExports)
	close(suite.stageRunner.ResultsChan)
}

func (suite *StageRunnerSuite) TestXunitParser(check *gocheck.C) {
	suite.stageRunner = New(suite.resourcesConnection, suite.virtualMachine,
		suite.build, new(MockNoExportsExporter))

	copyExecutable := shell.Executable{
		Command: shell.And(
			shell.Command("mkdir -p xuintPath/"),
			shell.Commandf("cp %s %s", path.Join(os.Getenv("GOPATH"), "src", "koality", "util", "xunitsamples", "*.xml"), "xuintPath/"),
		),
	}
	copyExecution, err := suite.virtualMachine.Execute(copyExecutable)
	check.Assert(err, gocheck.IsNil)
	err = copyExecution.Wait()
	check.Assert(err, gocheck.IsNil)

	compileCmd := exec.Command("go", "install", path.Join("koality", "util", "xunit", "getXunitResults"))
	err = compileCmd.Run()
	check.Assert(err, gocheck.IsNil)

	commands := []build.Command{
		remotecommand.NewRemoteCommand(false, ".", "command name!!", 1000, []string{"xunitPath"}, []string{"true"}),
	}
	for index, command := range commands {
		_, err := suite.resourcesConnection.Stages.Create.Create(suite.build.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			check.Check(result.Passed, gocheck.Equals, true, gocheck.Commentf("Failed section %s", result.Section))
		}

		doneChan <- err
	}(doneChan)

	err = suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	stage, err := suite.resourcesConnection.Stages.Read.GetBySectionNumberAndName(suite.build.Id, 0, "command name!!")
	check.Assert(err, gocheck.IsNil)

	stageRuns, err := suite.resourcesConnection.Stages.Read.GetAllRuns(stage.Id)
	check.Assert(err, gocheck.IsNil)
	check.Assert(stageRuns, gocheck.HasLen, 1)

	xunitResults, err := suite.resourcesConnection.Stages.Read.GetAllXunitResults(stageRuns[0].Id)
	check.Assert(err, gocheck.IsNil)

	xunitPath := path.Join(os.Getenv("GOPATH"), "src", "koality", "util", "xunitsamples")
	expectedXunitBytes, err := xunit.GetXunitResults("*.xml", []string{xunitPath}, ioutil.Discard, os.Stderr)
	check.Assert(err, gocheck.IsNil)

	var expectedXunitResults []resources.XunitResult
	err = json.Unmarshal(expectedXunitBytes, &expectedXunitResults)
	check.Assert(err, gocheck.IsNil)

	// normalize paths
	for i := range xunitResults {
		xunitResults[i].Path = path.Base(xunitResults[i].Path)
	}
	for i := range expectedXunitResults {
		expectedXunitResults[i].Path = path.Base(expectedXunitResults[i].Path)
	}
	check.Assert(len(xunitResults), gocheck.DeepEquals, len(expectedXunitResults))

	close(suite.stageRunner.ResultsChan)
}

func (suite *StageRunnerSuite) TestSimpleFailingStages(check *gocheck.C) {
	suite.stageRunner = New(suite.resourcesConnection, suite.virtualMachine,
		suite.build, new(MockNoExportsExporter))

	commands := []build.Command{build.NewShellCommand("fail", shell.Command("false"))}
	for index, command := range commands {
		_, err := suite.resourcesConnection.Stages.Create.Create(suite.build.Id, 0, command.Name(), uint64(index))
		check.Assert(err, gocheck.IsNil)
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range suite.stageRunner.ResultsChan {
			check.Check(result.Passed, gocheck.Equals, false, gocheck.Commentf("Passed section %s", result.Section))
		}

		doneChan <- nil
	}(doneChan)

	err := suite.stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	check.Assert(err, gocheck.IsNil)

	close(suite.stageRunner.ResultsChan)
}
