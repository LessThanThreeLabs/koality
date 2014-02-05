package debuginstancerunner

import (
	"github.com/LessThanThreeLabs/gocheck"
	"github.com/dchest/goyaml"
	"io/ioutil"
	"koality/repositorymanager"
	"koality/resources"
	"koality/resources/database"
	"koality/vm"
	"koality/vm/localmachine"
	"koality/vm/poolmanager"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type DebugInstanceRunnerSuite struct {
	resourcesConnection *resources.Connection
	repository          *resources.Repository
	verification        *resources.Verification
	debugInstance       *resources.DebugInstance
	debugInstanceRunner *DebugInstanceRunner
	repositoryManager   repositorymanager.RepositoryManager
	tmpDir              string
	repoPath            string
	instanceId          string
}

var (
	_              = gocheck.Suite(&DebugInstanceRunnerSuite{})
	anotherSection = map[string]interface{}{
		"another section": map[string]interface{}{
			"run on":  "split",
			"fail on": "any",
			"scripts": []interface{}{
				"true",
			},
		},
	}
	failingSection = map[string]interface{}{
		"a failing section": map[string]interface{}{
			"run on":  "split",
			"fail on": "any",
			"scripts": []interface{}{
				"foo",
			},
		},
	}
	passingSection = map[string]interface{}{
		"a passing section": map[string]interface{}{
			"run on":  "split",
			"fail on": "any",
			"scripts": []interface{}{
				"pwd",
			},
		},
	}
	finalSectionSlice = []interface{}{
		map[string]interface{}{
			"a final section": map[string]interface{}{
				"run on":  "single",
				"fail on": "any",
				"scripts": []interface{}{
					"echo $KOALITY_STATUS",
					"printenv",
					"false",
				},
			},
		},
	}
)

func (suite *DebugInstanceRunnerSuite) SetUpTest(check *gocheck.C) {
	err := database.PopulateDatabase()
	check.Assert(err, gocheck.IsNil)

	suite.resourcesConnection, err = database.New()
	check.Assert(err, gocheck.IsNil)

	suite.tmpDir, err = ioutil.TempDir("", "tmp@") // We need an @ symbol to beat the repositories remoteUri verifier
	check.Assert(err, gocheck.IsNil)

	suite.repoPath = path.Join(suite.tmpDir, "testRepo")
	err = exec.Command("git", "init", suite.repoPath).Run()
	check.Assert(err, gocheck.IsNil)
}

func (suite *DebugInstanceRunnerSuite) TearDownTest(check *gocheck.C) {
	if suite.resourcesConnection != nil {
		suite.resourcesConnection.Close()
	}

	if suite.tmpDir != "" {
		os.RemoveAll(suite.tmpDir)
	}

	if suite.repositoryManager != nil && suite.repository != nil {
		suite.repositoryManager.DeleteRepository(suite.repository)
	}
}

func (suite *DebugInstanceRunnerSuite) runDebugInstanceWithYaml(check *gocheck.C, yml map[string]interface{}) (bool, error) {
	ymlBytes, err := goyaml.Marshal(yml)
	check.Assert(err, gocheck.IsNil)

	err = ioutil.WriteFile(path.Join(suite.repoPath, "koality.yml"), ymlBytes, 0664)
	check.Assert(err, gocheck.IsNil)

	cmd := exec.Command("git", "add", "koality.yml")
	cmd.Dir = suite.repoPath
	output, err := cmd.CombinedOutput()
	check.Assert(err, gocheck.IsNil, gocheck.Commentf(string(output)))

	cmd = exec.Command("git", "commit", "koality.yml", "-m", "First commit", "--author", "Test User <test@us.er>")
	cmd.Dir = suite.repoPath
	output, err = cmd.CombinedOutput()
	check.Assert(err, gocheck.IsNil)

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = suite.repoPath
	shaBytes, err := cmd.Output()
	check.Assert(err, gocheck.IsNil)

	sha := strings.TrimSpace(string(shaBytes))

	err = os.Mkdir(path.Join(suite.repoPath, ".git", "refs", "koality"), 0777)
	check.Assert(err, gocheck.IsNil)

	err = ioutil.WriteFile(path.Join(suite.repoPath, ".git", repositorymanager.GitHiddenRef(sha)), shaBytes, 0664)
	check.Assert(err, gocheck.IsNil)

	suite.repository, err =
		suite.resourcesConnection.Repositories.Create.Create("repositoryName", "git", "file://"+suite.repoPath)
	check.Assert(err, gocheck.IsNil)

	suite.repositoryManager = repositorymanager.New("/tmp/repositoryManager", suite.resourcesConnection)
	err = suite.repositoryManager.CreateRepository(suite.repository)
	check.Assert(err, gocheck.IsNil)

	suite.instanceId = "identifier"
	vmPool := vm.NewPool(1, localmachine.Manager, 0, 3)
	poolManager := poolmanager.New([]vm.VirtualMachinePool{vmPool})

	debugInstanceRunner := New(suite.resourcesConnection, poolManager, suite.repositoryManager)
	expires := time.Now() // don't run the debug instance any longer after running the stages
	suite.debugInstance, err = suite.resourcesConnection.DebugInstances.Create.Create(
		vmPool.Id(), suite.instanceId, &expires, &resources.CoreVerificationInformation{
			suite.repository.Id, sha, "1234567890123456789012345678901234567890", "headMessage",
			"headUsername", "head@Ema.il", "a@b.com"})
	check.Assert(err, gocheck.IsNil)

	return debugInstanceRunner.RunDebugInstance(suite.debugInstance)
}

func (suite *DebugInstanceRunnerSuite) TestDebugInstanceRunsPassingStages(check *gocheck.C) {
	success, err := suite.runDebugInstanceWithYaml(check, map[string]interface{}{
		"parameters": map[string]interface{}{
			"nodes": 8,
		},
		"sections": []interface{}{passingSection},
		"final":    finalSectionSlice,
	})
	check.Assert(err, gocheck.IsNil)
	check.Assert(success, gocheck.Equals, true)
}

func (suite *DebugInstanceRunnerSuite) TestDebugInstanceRunsUpToNotIncludingSnapshotUntil2(check *gocheck.C) {
	success, err := suite.runDebugInstanceWithYaml(check, map[string]interface{}{
		"parameters": map[string]interface{}{
			"nodes":          8,
			"snapshot until": "another section",
		},
		"sections": []interface{}{passingSection, failingSection, anotherSection},
	})
	check.Assert(err, gocheck.IsNil)
	check.Assert(success, gocheck.Equals, false)
}

func (suite *DebugInstanceRunnerSuite) TestDebugInstanceRunsUpToNotIncludingSnapshotUntil(check *gocheck.C) {
	success, err := suite.runDebugInstanceWithYaml(check, map[string]interface{}{
		"parameters": map[string]interface{}{
			"nodes":          8,
			"snapshot until": "a failing section",
		},
		"sections": []interface{}{passingSection, failingSection, anotherSection},
	})
	check.Assert(err, gocheck.IsNil)
	check.Assert(success, gocheck.Equals, true)
}

func (suite *DebugInstanceRunnerSuite) TestDebugInstanceRunsNormalStages(check *gocheck.C) {
	success, err := suite.runDebugInstanceWithYaml(check, map[string]interface{}{
		"parameters": map[string]interface{}{
			"nodes": 8,
		},
		"sections": []interface{}{failingSection},
	})
	check.Assert(err, gocheck.IsNil)
	check.Assert(success, gocheck.Equals, false)
}

func (suite *DebugInstanceRunnerSuite) TestDebugInstanceRunsNormalStages2(check *gocheck.C) {
	success, err := suite.runDebugInstanceWithYaml(check, map[string]interface{}{
		"parameters": map[string]interface{}{
			"nodes": 8,
		},
		"sections": []interface{}{passingSection},
	})
	check.Assert(err, gocheck.IsNil)
	check.Assert(success, gocheck.Equals, true)
}

func (suite *DebugInstanceRunnerSuite) TestDebugInstanceDoesntRunFinalStages(check *gocheck.C) {
	success, err := suite.runDebugInstanceWithYaml(check, map[string]interface{}{
		"parameters": map[string]interface{}{
			"nodes": 8,
		},
		"sections": []interface{}{passingSection},
		"final":    finalSectionSlice,
	})
	check.Assert(err, gocheck.IsNil)
	check.Assert(success, gocheck.Equals, true)

	stages, err := suite.resourcesConnection.Stages.Read.GetAll(suite.debugInstance.VerificationId)
	check.Assert(err, gocheck.IsNil)
	check.Assert(stages, gocheck.HasLen, 3) // git, provision, pwd
}
