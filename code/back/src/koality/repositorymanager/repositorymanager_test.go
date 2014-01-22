package repositorymanager

import (
	"fmt"
	"io/ioutil"
	"koality/resources"
	"os"
	"path/filepath"
	"testing"
)

var buf []byte

var remoteRepositoryPath = filepath.Join("/", "etc", "koality", "repositories", "remote")
var RM = repositoryManager{filepath.Join("/", "etc", "koality")}

var (
	gitRepositoryResource = &resources.Repository{0, "gitRepository", "", "git", "", remoteRepositoryPath, nil, nil}
	gitRepo               = RM.openGitRepository(gitRepositoryResource)
	gitRemoteRepository   = &gitSubRepository{remoteRepositoryPath}
)

var (
	hgRepositoryResource = &resources.Repository{1, "hgRepository", "", "hg", "", remoteRepositoryPath, nil, nil}
	hgRepo               = RM.openHgRepository(hgRepositoryResource)
	hgRemoteRepository   = &hgRepository{remoteRepositoryPath, ""}
)

func getTopRef(remoteRepository Repository) (ref string) {
	if remoteRepository.getVcsBaseCommand() == "git" {
		ref = "HEAD"
	} else {
		ref = "tip"
	}

	return
}

func writeAdd(repository Repository, filename, filechange string) {
	ioutil.WriteFile(filepath.Join(repository.getPath(), filename), []byte(filechange), 0644)
	RunCommand(Command(repository, nil, "add", filename))

}

func repositoryTestSetup(repository StoredRepository, remoteRepository Repository, testing *testing.T) {
	os.MkdirAll(remoteRepositoryPath, 0700)
	RunCommand(Command(remoteRepository, nil, "init"))

	writeAdd(remoteRepository, "koality.yml", "example yaml file contents")

	if remoteRepository.getVcsBaseCommand() == "git" {
		RunCommand(Command(remoteRepository, nil, "commit", "-m", "Add koality.yml file", "--author=<chicken chickenson <cchickenson@chicken.com>>"))
	} else {
		RunCommand(Command(remoteRepository, nil, "commit", "-m", "Add koality.yml file", "-u", "chicken chickenson <cchickenson@chicken.com>"))
	}

	if err := repository.createRepository(); err != nil {
		testing.Fatal(err)
	}
}

func repositoryTestTeardown(repository StoredRepository, testing *testing.T) {
	if err := repository.deleteRepository(); err != nil {
		testing.Fatal(err)
	}

	os.RemoveAll(remoteRepositoryPath)
}

func testCreateGetYamlDelete(repository StoredRepository, remoteRepository Repository, testing *testing.T) {
	var err error
	defer func() {
		repositoryTestTeardown(repository, testing)
		if _, err = os.Stat(gitRepo.bare.path); !os.IsNotExist(err) {
			testing.Fatal("Repository still existed after deletion.")
		}
	}()

	repositoryTestSetup(repository, remoteRepository, testing)

	yamlFile, err := repository.getYamlFile(getTopRef(remoteRepository))
	if err != nil {
		testing.Fatal("Upon creation, repository did not clone properly, giving err=", err)
	} else if yamlFile != "example yaml file contents" {
		testing.Fatal("GetYamlFile did not return expected value for", remoteRepository.getVcsBaseCommand(), ".")
	}
}

func testGetCommitAttributes(repository StoredRepository, remoteRepository Repository, testing *testing.T) {
	defer repositoryTestTeardown(repository, testing)
	repositoryTestSetup(repository, remoteRepository, testing)

	message, username, email, err := repository.getCommitAttributes(getTopRef(remoteRepository))

	if err != nil {
		testing.Fatal(err)
	} else if message != "Add koality.yml file" || username != "chicken chickenson" || email != "cchickenson@chicken.com" {
		testing.Fatal("Getting commit attributes did not work as intended for", remoteRepository.getVcsBaseCommand(), ".")
	}
}

func TestGitStorePending(testing *testing.T) {
	defer repositoryTestTeardown(gitRepo, testing)
	repositoryTestSetup(gitRepo, gitRemoteRepository, testing)

	headSha, _ := gitRemoteRepository.getHeadSha()

	if err := gitRepo.storePending(headSha, remoteRepositoryPath); err != nil {
		testing.Fatal(err)
	}

	if err := RunCommand(Command(gitRemoteRepository, nil, "show", fmt.Sprintf("refs/koality/%s", headSha))); err != nil {
		testing.Fatal(err)
	}
}

var (
	//The cloned repository is the repository that would be pushing to the local bare repository
	clonedRepositoryPath = filepath.Join("/", "etc", "koality", "repositories", "clone")
	clonedRepository     = &gitSubRepository{clonedRepositoryPath}
)

func TestGitMergePass(testing *testing.T) {
	defer repositoryTestTeardown(gitRepo, testing)
	defer os.RemoveAll(clonedRepositoryPath)
	repositoryTestSetup(gitRepo, gitRemoteRepository, testing)

	os.MkdirAll(clonedRepositoryPath, 0700)
	RunCommand(Command(clonedRepository, nil, "clone", gitRepo.bare.path, clonedRepositoryPath))

	oldHeadSha, _ := clonedRepository.getHeadSha()

	writeAdd(clonedRepository, "newfile", "test changeset")
	RunCommand(Command(clonedRepository, nil, "commit", "-m", "New commit", "--author=<chicken chickenson <cchickenson@chicken.com>>"))

	headSha, _ := clonedRepository.getHeadSha()

	if err := RunCommand(Command(clonedRepository, nil, "push", "origin", fmt.Sprintf("HEAD:refs/for/%s", headSha))); err != nil {
		fmt.Println(err)
	}

	if err := RM.MergeChangeset(gitRepositoryResource, fmt.Sprintf("refs/for/%s", headSha), fmt.Sprintf("refs/for/%s", headSha), oldHeadSha); err != nil {
		testing.Fatal(err)
	}

	if _, _, _, err := gitRepo.getCommitAttributes(headSha); err != nil {
		testing.Fatal("Merging did not result in the new commit being present in the main repository.")
	}
}

//TODO(akostov) Talk to Jon + add more git merging tests.

func TestGitCreateGetYamlDelete(testing *testing.T) {
	testCreateGetYamlDelete(gitRepo, gitRemoteRepository, testing)
}

func TestHgCreateGetYamlDelete(testing *testing.T) {
	testCreateGetYamlDelete(hgRepo, hgRemoteRepository, testing)
}

func TestGitGetCommitAttributes(testing *testing.T) {
	testGetCommitAttributes(gitRepo, gitRemoteRepository, testing)
}

func TestHgGetCommitAttributes(testing *testing.T) {
	testGetCommitAttributes(hgRepo, hgRemoteRepository, testing)
}
