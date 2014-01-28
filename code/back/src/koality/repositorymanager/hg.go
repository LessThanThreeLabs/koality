package repositorymanager

import (
	"fmt"
	"koality/resources"
	"koality/shell"
	"os"
	"strings"
)

type hgRepository struct {
	path                string
	remoteUri           string
	resourcesConnection *resources.Connection
}

func (repositoryManager *repositoryManager) openHgRepository(repository *resources.Repository) *hgRepository {
	return &hgRepository{repositoryManager.ToPath(repository), repository.RemoteUri, repositoryManager.resourcesConnection}
}

func (repository *hgRepository) getVcsBaseCommand() string {
	return "hg"
}

func (repository *hgRepository) getPath() string {
	return repository.path
}

func (repository *hgRepository) fetchWithPrivateKey(args ...string) (err error) {
	keyPair, err := repository.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		return
	}

	if err := RunCommand(Command(repository, nil, "pull", append([]string{"--ssh", shell.Quote(fmt.Sprintf("SSH_PRIVATE_KEY=%s %s -o ConnectTimeout=%s", keyPair.PrivateKey, defaultSshScript, defaultTimeout)), repository.remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func (repository *hgRepository) createRepository() (err error) {
	if _, err = os.Stat(repository.path); !os.IsNotExist(err) {
		return RepositoryAlreadyExistsInStoreError{fmt.Sprintf("The repository at %s already exists in the repository store.", repository.path)}
	}

	if err = os.MkdirAll(repository.path, 0700); err != nil {
		return
	}

	if err := RunCommand(Command(repository, nil, "init")); err != nil {
		return err
	}

	if err = repository.fetchWithPrivateKey(); err != nil {
		return
	}

	return
}

func (repository *hgRepository) deleteRepository() (err error) {
	if err = checkRepositoryExists(repository.path); err != nil {
		return err
	}

	return os.RemoveAll(repository.path)
}

func (repository *hgRepository) getTopSha(ref string) (topSha string, err error) {
	showCommand := Command(repository, nil, "log", "-r", ref)
	if err = RunCommand(showCommand); err != nil {
		return
	}

	shaLine, err := showCommand.Stdout.ReadString('\n')
	if err != nil {
		return
	}

	if !strings.HasPrefix(shaLine, "changeset:") {
		err = fmt.Errorf("git show %s output data for repository at %s was not formatted as expected.", ref, repository.path)
		return
	}

	topSha = strings.TrimSpace(strings.TrimPrefix(shaLine, "changeset:"))
	return
}

func (repository *hgRepository) getCommitAttributes(ref string) (headSha, message, username, email string, err error) {
	command := Command(repository, nil, "log", "-r", ref)
	if err = RunCommand(command); err != nil {
		err = NoSuchCommitInRepositoryError{fmt.Sprintf(fmt.Sprintf("The repository %v does not contain commit %s", repository, ref))}
		return
	}

	shaLine, err := command.Stdout.ReadString('\n')
	if err != nil {
		return
	}

	if !strings.HasPrefix(shaLine, "changeset:") {
		err = fmt.Errorf("hg log -r %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	tagLine, err := command.Stdout.ReadString('\n')

	if !strings.HasPrefix(tagLine, "tag: ") {
		err = fmt.Errorf("hg log -r %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	authorLine, err := command.Stdout.ReadString('\n')

	if !strings.HasPrefix(authorLine, "user: ") {
		err = fmt.Errorf("hg log -r %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	author := strings.TrimPrefix(authorLine, "user: ")

	authorSplit := strings.Split(strings.TrimSpace(author), " <")

	username = strings.TrimSpace(authorSplit[0])
	email = strings.Trim(strings.TrimSpace(authorSplit[1]), ">")

	dateLine, err := command.Stdout.ReadString('\n')

	if !strings.HasPrefix(dateLine, "date:") {
		err = fmt.Errorf("hg log -r %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	messageLine, err := command.Stdout.ReadString('\n')

	if !strings.HasPrefix(messageLine, "summary:") {
		err = fmt.Errorf("hg log -r %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	message = strings.TrimSpace(strings.TrimPrefix(messageLine, "summary:"))

	return
}

func (repository *hgRepository) getYamlFile(ref string) (yamlFile string, err error) {
	command := Command(repository, nil, "cat", "-r", ref, "koality.yml")
	if err = RunCommand(command); err != nil {
		return
	}

	yamlFile = command.Stdout.String()
	return
}
