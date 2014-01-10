package repositorystore

import (
	"fmt"
	"koality/repositorymanager/pathgenerator"
	"koality/resources"
	"os"
	"strings"
	"time"
)

const (
	pushMergeRetries = 4
	retryTimeout     = 1000000000 // In nanoseconds
)

type GitSubRepository struct {
	path string
}

func (repository *GitSubRepository) getVcsBaseCommand() string {
	return "git"
}

func (repository *GitSubRepository) getPath() string {
	return repository.path
}

type GitRepository struct {
	bare  *GitSubRepository
	slave *GitSubRepository

	remoteUri string
}

func OpenGitRepository(repository *resources.Repository) *GitRepository {
	path := pathgenerator.ToPath(repository)
	return &GitRepository{&GitSubRepository{path}, &GitSubRepository{path + ".slave"}, repository.RemoteUri}
}

func (repository *GitSubRepository) fetchWithPrivateKey(remoteUri string, args ...string) (err error) {
	env := []string{
		fmt.Sprintf("GIT_SSH=%s", defaultSshScript),
		fmt.Sprintf("GIT_PRIVATE_KEY_PATH=%s", defaultPrivateKeyPath),
		fmt.Sprintf("GIT_SSH_TIMEOUT=%s", defaultTimeout),
	}

	if err := RunCommand(Command(repository, env, "remote", "prune", remoteUri)); err != nil {
		return err
	}

	if err := RunCommand(Command(repository, env, "fetch", append([]string{remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func (repository *GitSubRepository) pushWithPrivateKey(remoteUri string, args ...string) (err error) {
	env := []string{
		fmt.Sprintf("GIT_SSH=%s", defaultSshScript),
		fmt.Sprintf("GIT_PRIVATE_KEY_PATH=%s", defaultPrivateKeyPath),
		fmt.Sprintf("GIT_SSH_TIMEOUT=%s", defaultTimeout),
	}

	if err := RunCommand(Command(repository, env, "push", append([]string{remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func (repository *GitRepository) StorePending(ref, remoteUri string, args ...string) (err error) {
	if err = repository.bare.fetchWithPrivateKey(remoteUri, "+refs/*:refs/*"); err != nil {
		return
	}

	if err = RunCommand(Command(repository.bare, nil, "show", ref)); err != nil {
		return NoSuchCommitInRepositoryError{fmt.Sprintf("The repository at %s does not contain commit %s", repository.bare.path, ref)}
	}

	if err = repository.bare.pushWithPrivateKey(remoteUri, fmt.Sprintf("%s:refs/pending/%s", ref, ref)); err != nil {
		return
	}

	return
}

func (repository *GitRepository) CreateRepository() (err error) {
	if err = checkRepositoryExists(repository.bare.path); err == nil {
		return
	}

	if _, err = os.Stat(repository.bare.path); !os.IsNotExist(err) {
		return RepositoryAlreadyExistsInStoreError{fmt.Sprintf("The repository at %s already exists in the repository store.", repository.bare.path)}
	}

	if err = os.MkdirAll(repository.bare.path, 0700); err != nil {
		return
	}

	if err := RunCommand(Command(repository.bare, nil, "init", "--bare")); err != nil {
		return err
	}

	if err = repository.bare.fetchWithPrivateKey(repository.remoteUri, "+refs/heads/*:refs/heads/*"); err != nil {
		return
	}

	if err = RunCommand(Command(repository.bare, nil, "clone", repository.bare.path, repository.slave.path)); err != nil {
		return err
	}

	return
}

func (repository *GitRepository) DeleteRepository() (err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return err
	}

	err = os.RemoveAll(repository.slave.path)
	if err != nil {
		return err
	}

	return os.RemoveAll(repository.bare.path)
}

func (repository *GitRepository) MergeChangeset(headRef, baseRef, refToMergeInto string) (err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return
	}

	originalHead, err := repository.mergeRefs(headRef, refToMergeInto)
	if err != nil {
		return
	}

	if err = repository.pushMergeRetry(repository.remoteUri, refToMergeInto, originalHead); err != nil {
		return
	}

	return
}

func (repository *GitRepository) mergeRefs(refToMerge, refToMergeInto string) (originalHead string, err error) {
	defer RunCommand(Command(repository.slave, nil, "reset", "--hard"))

	if err = RunCommand(Command(repository.slave, nil, "remote", "prune", "origin")); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "fetch")); err != nil {
		return
	}

	remoteBranch := fmt.Sprintf("origin/%s", refToMergeInto)
	branchCommand := Command(repository.slave, nil, "branch", "-r")
	if err = RunCommand(branchCommand); err != nil {
		return
	}

	branches := branchCommand.Stdout.String()
	var checkoutBranch string
	if remoteBranchExists := strings.Contains(branches, fmt.Sprintf("\\s+%s$", remoteBranch)); remoteBranchExists {
		checkoutBranch = remoteBranch
	} else {
		checkoutBranch = "FETCH_HEAD"
	}

	if err = RunCommand(Command(repository.slave, nil, "fetch", "origin", refToMerge)); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "checkout", "FETCH_HEAD")); err != nil {
		return
	}

	refSha, err := repository.slave.getHeadSha()
	if err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "branch", "-f", refToMergeInto, checkoutBranch)); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "checkout", refToMergeInto)); err != nil {
		return
	}

	if originalHead, err = repository.slave.getHeadSha(); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "merge", "FETCH_HEAD", "-m", fmt.Sprintf("Merging in %s", refSha))); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "push", "origin", fmt.Sprintf("HEAD:%s", refToMergeInto))); err != nil {
		return
	}

	return
}

func (repository *GitSubRepository) getHeadSha() (headSha string, err error) {
	showCommand := Command(repository, nil, "show", "HEAD")
	if err = RunCommand(showCommand); err != nil {
		return
	}

	shaLine, err := showCommand.Stdout.ReadString('\n')
	if err != nil {
		return
	}

	if !strings.HasPrefix(shaLine, "commit ") {
		err = fmt.Errorf("git show HEAD output data for repository at %s was not formatted as expected.", repository.path)
		return
	}

	headSha = strings.TrimSpace(strings.TrimPrefix(shaLine, "commit "))
	return
}

func (repository *GitSubRepository) resetRepositoryHead(refToReset, originalHead string) error {
	return RunCommand(Command(repository, nil, "push", "origin", fmt.Sprintf("%s:%s", originalHead, refToReset), "--force"))
}

func (repository *GitSubRepository) updateBranchFromForwardUrl(remoteUri, refToUpdate string) (headSha string, err error) {
	remoteBranch := fmt.Sprintf("origin/%s", refToUpdate)

	if err = repository.fetchWithPrivateKey(remoteUri, refToUpdate); err != nil {
		return
	}

	if err = RunCommand(Command(repository, nil, "checkout", "FETCH_HEAD")); err != nil {
		return
	}

	if headSha, err = repository.getHeadSha(); err != nil {
		return
	}

	if err = RunCommand(Command(repository, nil, "branch", "-f", refToUpdate, remoteBranch)); err != nil {
		return
	}

	if err = RunCommand(Command(repository, nil, "checkout", refToUpdate)); err != nil {
		return
	}

	return
}

func (repository *GitRepository) updateFromForwardUrl(remoteUri, refToMergeInto, originalHead string) (err error) {
	defer func() {
		if err != nil {
			repository.slave.resetRepositoryHead(refToMergeInto, originalHead)
		} else {
			RunCommand(Command(repository.slave, nil, "reset", "--hard"))
		}
	}()

	refSha, err := repository.slave.updateBranchFromForwardUrl(remoteUri, refToMergeInto)
	if err != nil {
		return
	}

	remoteBranch := fmt.Sprintf("origin/%s", refToMergeInto)

	if err = RunCommand(Command(repository.slave, nil, "checkout", remoteBranch)); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "branch", "-f", refToMergeInto, remoteBranch)); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "checkout", refToMergeInto)); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "merge", "FETCH_HEAD", "-m", fmt.Sprintf("Merging in %s", refSha))); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "push", "origin", fmt.Sprintf("HEAD:%s", refToMergeInto))); err != nil {
		return
	}

	return
}

func (repository *GitRepository) pushMergeRetry(remoteUri, refToMergeInto, originalHead string) (err error) {
	pushAttempts := 0

	for {
		pushAttempts += 1
		err = repository.bare.pushWithPrivateKey(remoteUri, fmt.Sprintf("%s:%s", refToMergeInto, refToMergeInto))

		//TODO(akostov) More precise error catching
		if err != nil && pushAttempts < pushMergeRetries {
			time.Sleep(retryTimeout)
			repository.updateFromForwardUrl(remoteUri, refToMergeInto, originalHead)
		} else if err != nil {
			repository.slave.resetRepositoryHead(refToMergeInto, originalHead)
		} else {
			break
		}
	}
	return
}

func (repository *GitRepository) GetCommitAttributes(ref string) (message, username, email string, err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return
	}

	command := Command(repository.bare, nil, "show", ref)
	if err = RunCommand(command); err != nil {
		err = NoSuchCommitInRepositoryError{fmt.Sprintf(fmt.Sprintf("The repository %v does not contain commit %s", repository, ref))}
		return
	}

	shaLine, err := command.Stdout.ReadString('\n')
	if err != nil {
		return
	}

	if !strings.HasPrefix(shaLine, "commit ") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	authorLine, err := command.Stdout.ReadString('\n')

	if !strings.HasPrefix(authorLine, "Author") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	author := strings.TrimSpace(strings.TrimPrefix(authorLine, "Author: "))

	authorSplit := strings.Split(author, " <")

	username = strings.TrimSpace(authorSplit[0])
	email = strings.Trim(strings.TrimSpace(authorSplit[1]), ">")

	dateLine, err := command.Stdout.ReadString('\n')

	if !strings.HasPrefix(dateLine, "Date:") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	blankLine, err := command.Stdout.ReadString('\n')

	if blankLine != "\n" {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	messageLine, err := command.Stdout.ReadString('\n')

	message = strings.TrimSpace(messageLine)

	return
}

func (repository *GitRepository) GetYamlFile(ref string) (yamlFile string, err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return
	}

	// TODO(akostov) discuss getting rid of .koality.yml file
	command := Command(repository.bare, nil, "show", fmt.Sprintf("%s:koality.yml", ref))
	if err = RunCommand(command); err != nil {
		return
	}

	yamlFile = command.Stdout.String()
	return
}
