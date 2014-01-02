package repositorystore

import (
	"fmt"
	"koality/repositorymanager/pathgenerator"
	"koality/resources"
	"os"
)

func hgFetchWithPrivateKey(repository *Repository, remoteUri string, args ...string) (err error) {
	//TODO(akostov) GIT_PRIVATE_KEY_PATH = change script much
	if err := RunCommand(repository.Command(nil, "pull", append([]string{"--ssh ", fmt.Sprintf("\"GIT_PRIVATE_KEY_PATH=%s %s -o ConnectTimeout=%s\"", defaultPrivateKeyPath, defaultSshScript, defaultTimeout), remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func hgCreateRepository(repository *resources.Repository) (err error) {
	path := pathgenerator.ToPath(repository)
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		return RepositoryAlreadyExistsInStoreError{fmt.Sprintf("The repository %v already exists in the repository store.", repository.Name)}
	}

	if err = os.MkdirAll(path, 0700); err != nil {
		return
	}

	localRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return err
	}

	if err := RunCommand(localRepository.Command(nil, "init")); err != nil {
		return err
	}

	if err = hgFetchWithPrivateKey(localRepository, repository.RemoteUri); err != nil {
		return
	}

	return
}

func hgGetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return
	}

	storedRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return
	}

	command := storedRepository.Command(nil, "log", "-r", ref)
	if err = RunCommand(command); err != nil {
		err = NoSuchCommitInRepositoryError{fmt.Sprintf(fmt.Sprintf("The repository %v does not contain commit %s", repository, ref))}
		return
	}

	commitDataReader := bufio.NewReader(strings.NewReader(command.Stdout.String()))

	shaLine, err := commitDataReader.ReadString('\n')
	if err != nil {
		return
	}

	if !strings.HasPrefix(shaLine, "changeset:") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	authorLine, err := commitDataReader.ReadString('\n')

	author := strings.TrimPrefix(authorLine, "user: ")
	if author == authorLine {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	authorSplit := strings.Split(strings.Trim(author, " "), " <")

	username = strings.Trim(authorSplit[0], " ")
	email = strings.Trim(authorSplit[1], "> ")

	dateLine, err := commitDataReader.ReadString('\n')

	if !strings.HasPrefix(dateLine, "date:") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	messageLine, err := commitDataReader.ReadString('\n')

	if !strings.HasPrefix("summary:") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	message = strings.Trim(strings.TrimPrefix(messageLine, "summary:"), " ")

	return
}

func hgGetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return
	}

	storedRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return
	}

	command := storedRepository.Command(nil, "cat", ref, "koality.yml")
	if err = RunCommand(command); err != nil {
		return
	}

	yamlFile = command.Stdout.String()
	return
}
