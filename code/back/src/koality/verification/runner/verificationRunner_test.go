package runner

import (
	"github.com/dchest/goyaml"
	"io/ioutil"
	"koality/repositorymanager"
	"koality/resources/database"
	"koality/shell"
	"koality/vm"
	"koality/vm/localmachine"
	"koality/vm/poolmanager"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func testVerification(test *testing.T, ymlBytes []byte, expectSuccess bool) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer resourcesConnection.Close()

	tmpDir, err := ioutil.TempDir("", "tmp@") // We need an @ symbol to beat the repositories remoteUri verifier
	defer os.RemoveAll(tmpDir)

	repoPath := path.Join(tmpDir, "testRepo")

	err = exec.Command("git", "init", repoPath).Run()
	if err != nil {
		test.Fatal(err)
	}

	err = ioutil.WriteFile(path.Join(repoPath, "koality.yml"), ymlBytes, 0664)
	if err != nil {
		test.Fatal(err)
	}

	cmd := exec.Command("git", "add", "koality.yml")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		test.Fatal(err, string(output))
	}

	cmd = exec.Command("git", "commit", "koality.yml", "-m", "First commit", "--author", "Test User <test@us.er>")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		test.Fatal(err, string(output))
	}

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	shaBytes, err := cmd.Output()
	if err != nil {
		test.Fatal(err, shaBytes)
	}

	sha := strings.TrimSpace(string(shaBytes))

	err = os.Mkdir(path.Join(repoPath, ".git", "refs", "koality"), 0777)
	if err != nil {
		test.Fatal(err)
	}

	err = ioutil.WriteFile(path.Join(repoPath, ".git", repositorymanager.GitHiddenRef(sha)), shaBytes, 0664)
	if err != nil {
		test.Fatal(err)
	}

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "file://"+repoPath)
	if err != nil {
		test.Fatal(err)
	}

	repositoryManager := repositorymanager.New("/tmp/repositoryManager")

	err = repositoryManager.CreateRepository(repository)
	if err != nil {
		test.Fatal(err)
	}
	defer os.RemoveAll(path.Dir(repositoryManager.ToPath(repository)))

	vmPool := vm.NewPool(0, localmachine.Manager, 0, 3)
	poolManager := poolmanager.New([]vm.VirtualMachinePool{vmPool})

	verificationRunner := New(resourcesConnection, poolManager, repositoryManager)

	verification, err := resourcesConnection.Verifications.Create.Create(repository.Id, sha, "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	success, err := verificationRunner.RunVerification(verification)
	if err != nil {
		test.Fatal(err)
	}

	if !success && expectSuccess {
		test.Fatal("Verification failed, expected success")
	} else if success && !expectSuccess {
		test.Fatal("Verification passed, expected failure")
	}
}

func TestSimplePassingVerification(test *testing.T) {
	ymlBytes, err := goyaml.Marshal(
		map[string]interface{}{
			"parameters": map[string]interface{}{
				"nodes": 8,
			},
			"sections": []interface{}{
				map[string]interface{}{
					"a passing section": map[string]interface{}{
						"run on":  "split",
						"fail on": "any",
						"scripts": []interface{}{
							"pwd",
							"pwd",
							"pwd",
						},
					},
				},
			},
			"final": []interface{}{
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
			},
		},
	)
	if err != nil {
		test.Fatal(err)
	}

	testVerification(test, ymlBytes, true)
}

func TestSimpleFailingVerification(test *testing.T) {
	ymlBytes, err := goyaml.Marshal(
		map[string]interface{}{
			"parameters": map[string]interface{}{
				"nodes": 8,
			},
			"sections": []interface{}{
				map[string]interface{}{
					"a failing section": map[string]interface{}{
						"run on":  "split",
						"fail on": "any",
						"scripts": []interface{}{
							"pwd",
							"false",
							"pwd",
						},
					},
				},
			},
		},
	)
	if err != nil {
		test.Fatal(err)
	}

	testVerification(test, ymlBytes, false)
}

func TestEnvironment(test *testing.T) {
	const environmentVariableName = "ThisIsAnEnvironmentVariableName"
	const environmentVariableValue = "This is the environemnt variable value."
	ymlBytes, err := goyaml.Marshal(
		map[string]interface{}{
			"parameters": map[string]interface{}{
				"nodes": 1,
				"environment": map[string]string{
					environmentVariableName: environmentVariableValue,
				},
			},
			"sections": []interface{}{
				map[string]interface{}{
					"a passing section": map[string]interface{}{
						"run on":  "split",
						"fail on": "any",
						"scripts": []interface{}{
							shell.Test(shell.Commandf("\"$%s\" == %s", environmentVariableName, shell.Quote(environmentVariableValue))),
						},
					},
				},
			},
		},
	)
	if err != nil {
		test.Fatal(err)
	}

	testVerification(test, ymlBytes, true)
}
