package repositorymanager

import (
	"fmt"
	"koality/repositorymanager/pathgenerator"
	"koality/resources"
	"koality/shell"
	"os"
	"strings"
)

type HgRepository struct {
	path      string
	remoteUri string
}

func openHgRepository(repository *resources.Repository) *HgRepository {
	return &HgRepository{pathgenerator.ToPath(repository), repository.RemoteUri}
}

func (repository *HgRepository) getVcsBaseCommand() string {
	return "hg"
}

func (repository *HgRepository) getPath() string {
	return repository.path
}

func (repository *HgRepository) fetchWithPrivateKey(args ...string) (err error) {
	//TODO(akostov) GIT_PRIVATE_KEY_PATH = change script much
	if err := RunCommand(Command(repository, nil, "pull", append([]string{"--ssh", shell.Quote(fmt.Sprintf("GIT_PRIVATE_KEY_PATH=%s %s -o ConnectTimeout=%s", defaultPrivateKeyPath, defaultSshScript, defaultTimeout)), repository.remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func (repository *HgRepository) createRepository() (err error) {
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

func (repository *HgRepository) deleteRepository() (err error) {
	if err = checkRepositoryExists(repository.path); err != nil {
		return err
	}

	return os.RemoveAll(repository.path)
}

func (repository *HgRepository) getCommitAttributes(ref string) (message, username, email string, err error) {
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

func (repository *HgRepository) getYamlFile(ref string) (yamlFile string, err error) {
	command := Command(repository, nil, "cat", "-r", ref, "koality.yml")
	if err = RunCommand(command); err != nil {
		return
	}

	yamlFile = command.Stdout.String()
	return
}
