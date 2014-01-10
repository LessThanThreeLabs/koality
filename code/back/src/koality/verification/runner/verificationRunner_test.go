package runner

import (
	"github.com/dchest/goyaml"
	"io/ioutil"
	"koality/repositorymanager/pathgenerator"
	"koality/repositorymanager/repositorystore"
	"koality/resources/database"
	"koality/vm"
	"koality/vm/localmachine"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestSimplePassingVerification(test *testing.T) {
	database.PopulateDatabase()

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir("", "tmp@") // We need an @ symbol to beat the repositories remoteUri verifier
	defer os.RemoveAll(tmpDir)

	repoPath := path.Join(tmpDir, "testRepo")

	err = exec.Command("git", "init", repoPath).Run()
	if err != nil {
		test.Fatal(err)
	}

	ymlBytes, err := goyaml.Marshal(map[string]interface{}{"parameters": map[string]interface{}{"languages": map[string]interface{}{"python": 2.7}}})
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

	err = os.Mkdir(path.Join(repoPath, ".git", "refs", "pending"), 0777)
	if err != nil {
		test.Fatal(err)
	}

	err = ioutil.WriteFile(path.Join(repoPath, ".git", "refs", "pending", sha), shaBytes, 0664)
	if err != nil {
		test.Fatal(err)
	}

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "file://"+repoPath)
	if err != nil {
		test.Fatal(err)
	}

	gitRepository := repositorystore.OpenGitRepository(repository)

	err = gitRepository.CreateRepository()
	if err != nil {
		test.Fatal(err)
	}
	defer os.RemoveAll(path.Dir(pathgenerator.ToPath(repository)))

	vmPool := vm.NewPool(0, localmachine.NewLauncher(), 0, 3)

	verificationRunner := New(resourcesConnection, []vm.VirtualMachinePool{vmPool}, nil)

	verification, err := resourcesConnection.Verifications.Create.Create(repository.Id, sha, "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	success, err := verificationRunner.RunVerification(verification)
	if err != nil {
		test.Fatal(err)
	}

	if !success {
		test.Fatal("Verification failed")
	}
}
