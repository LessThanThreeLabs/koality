package repositorystore

import (
	"fmt"
	"koality/repositorymanager/pathgenerator"
	"koality/resources"
	"os"
)

const (
	defaultSshScript      = ""
	defaultPrivateKeyPath = ""
	defaultTimeout        = 120
)

func gitFetchWithPrivateKey(repository *Repository, remoteUri string, args ...string) (err error) {
	env := []string{
		fmt.Sprintf("GIT_SSH=%s", defaultSshScript),
		fmt.Sprintf("GIT_PRIVATE_KEY_PATH=%s", defaultPrivateKeyPath),
		fmt.Sprintf("GIT_SSH_TIMEOUT=%s", defaultTimeout),
	}

	if err := RunCommand(repository.Command(env, "remote", "prune")); err != nil {
		return err
	}

	if err := RunCommand(repository.Command(env, "fetch", append([]string{remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func gitPushWithPrivateKey(repository *resources.Repository, ref string) (err error) {
	return
}

func gitCreateRepository(repository *resources.Repository) (err error) {
	path := pathgenerator.ToPath(repository)
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		return RepositoryAlreadyExistsInStoreError{fmt.Sprintf("The repository %v already exists in the repository store.", repository.Name)}
	}

	if err = os.MkdirAll(path, 0700); err != nil {
		return
	}

	bareRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return err
	}

	if err := RunCommand(bareRepository.Command(nil, "init", "--bare")); err != nil {
		return err
	}

	if err = gitFetchWithPrivateKey(bareRepository, repository.RemoteUri, "+refs/heads/*:refs/heads/*"); err != nil {
		return
	}

	if err = RunCommand(bareRepository.Command(nil, "clone", path, path+".slave")); err != nil {
		return err
	}

	return
}

func gitMergeChangeset(repository *resources.Repository, headRef, baseRef, mergeIntoRef string) (err error) {
	return
}

func gitGetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	return
}

func gitGetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return "", err
	}

	gitRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return "", err
	}

	// TODO(akostov) .koality.yml file?
	command := gitRepository.Command(nil, "show", "koality.yml")
	if err := RunCommand(command); err != nil {
		return "", err
	}

	yamlFile = command.Stdout.String()
	return
}
