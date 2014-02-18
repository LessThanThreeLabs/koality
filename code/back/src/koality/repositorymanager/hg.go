package repositorymanager

import (
	"fmt"
	"koality/resources"
	"koality/shell"
	"os"
	"strings"
)

type hgRepository struct {
	name                string
	path                string
	remoteUri           string
	resourcesConnection *resources.Connection
}

var ensureHgInstalledCommand = shell.And(
	shell.Or(
		shell.Silent("which ssh"),
		shell.Advertised(shell.Sudo("apt-get install -y ssh-client")),
	),
	shell.Or(
		shell.Silent("which hg"),
		shell.Advertised(shell.Sudo("apt-get install -y mercurial")),
	),
)

func (repositoryManager *repositoryManager) openHgRepository(repository *resources.Repository) *hgRepository {
	return &hgRepository{repository.Name, repositoryManager.ToPath(repository), repository.RemoteUri, repositoryManager.resourcesConnection}
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

	if err := RunCommand(Command(repository, nil, "pull", append([]string{"--ssh", shell.Quote(fmt.Sprintf("PRIVATE_KEY=%s %s -o ConnectTimeout=%s", keyPair.PrivateKey, defaultSshScript, defaultTimeout)), repository.remoteUri}, args...)...)); err != nil {
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
	if err = checkRepositoryExists(repository.path); err != nil {
		return
	}

	if err = repository.fetchWithPrivateKey(); err != nil {
		return
	}

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
	if err = checkRepositoryExists(repository.path); err != nil {
		return
	}

	if err = repository.fetchWithPrivateKey(); err != nil {
		return
	}

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
	if err = checkRepositoryExists(repository.path); err != nil {
		return
	}

	command := Command(repository, nil, "cat", "-r", ref, "koality.yml")
	if err = RunCommand(command); err != nil {
		return
	}

	yamlFile = command.Stdout.String()
	return
}

func (repository *hgRepository) getCloneCommand() shell.Command {
	return shell.And(
		ensureHgInstalledCommand,
		shell.Or(
			shell.Not(shell.Test(shell.Commandf("-d %s", repository.name))),
			shell.Advertised(shell.Commandf("rm -rf %s", repository.name)),
		),
		shell.Advertised(shell.Commandf("hg clone --uncompressed %s %s", repository.remoteUri, repository.name)),
	)
}

func (repository *hgRepository) getCheckoutCommand(ref, baseRef string) shell.Command {
	commands := []shell.Command{
		ensureHgInstalledCommand,
		shell.IfElse(
			shell.Test(shell.Commandf("-d %s", repository.name)),
			shell.And(
				shell.Advertised(shell.Commandf("cd %s", repository.name)),
				shell.Advertised(shell.Commandf("hg pull %s", repository.remoteUri)),
			),
			shell.Chain(
				shell.Advertised(shell.Commandf("hg clone --uncompressed %s %s", repository.remoteUri, repository.name)),
				shell.Advertised(shell.Commandf("cd %s", repository.name)),
			),
		),
		shell.Advertised(shell.Commandf("hg update --clean %s", ref)),
	}
	if baseRef != ref && baseRef != "" {
		commands = append(commands, shell.Advertised(shell.Commandf("hg merge -y")))
	}
	return shell.And(commands...)
}
