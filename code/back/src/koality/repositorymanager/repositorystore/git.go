package repositorystore

import (
	"bufio"
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

func gitPushWithPrivateKey(repository *Repository, remoteUri string, args ...string) (err error) {
	env := []string{
		fmt.Sprintf("GIT_SSH=%s", defaultSshScript),
		fmt.Sprintf("GIT_PRIVATE_KEY_PATH=%s", defaultPrivateKeyPath),
		fmt.Sprintf("GIT_SSH_TIMEOUT=%s", defaultTimeout),
	}

	if err := RunCommand(repository.Command(env, "push", append([]string{remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func gitStorePending(repository *Repository, ref, remoteUri string, args ...string) (err error) {
	if err = gitFetchWithPrivateKey(repository, remoteUri, "+refs/*:refs/*"); err != nil {
		return
	}

	if err = RunCommand(repository.Command(nil, "show", ref)); err != nil {
		return NoSuchCommitInRepositoryError{fmt.Sprintf("The repository at %s does not contain commit %s", repository.path, ref)}
	}

	if err = gitPushWithPrivateKey(repository, remoteUri, fmt.Sprintf("%s:refs/pending/%s", ref, ref)); err != nil {
		return
	}

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

func gitMergeChangeset(repository *resources.Repository, headRef, baseRef, refToMergeInto string) (err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return
	}

	bareRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return
	}

	slaveRepository, err := Open(repository.VcsType, path+".slave")
	if err != nil {
		return
	}

	originalHead, err := gitMergeRefs(slaveRepository, headRef, refToMergeInto)
	if err != nil {
		return
	}

	if err = gitPushMergeRetry(bareRepository, slaveRepository, repository.RemoteUri, refToMergeInto, originalHead); err != nil {
		return
	}

	return
}

func gitMergeRefs(slaveRepository *Repository, refToMerge, refToMergeInto string) (originalHead string, err error) {
	defer RunCommand(slaveRepository.Command(nil, "reset", "--hard"))

	if err = RunCommand(slaveRepository.Command(nil, "remote", "prune", "origin")); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "fetch")); err != nil {
		return
	}

	remoteBranch := fmt.Sprintf("origin/%s", refToMergeInto)
	branchCommand := slaveRepository.Command(nil, "branch", "-r")
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

	if err = RunCommand(slaveRepository.Command(nil, "fetch", "origin", refToMerge)); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "checkout", "FETCH_HEAD")); err != nil {
		return
	}

	refSha, err := gitGetHeadSha(slaveRepository)
	if err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "branch", "-f", refToMergeInto, checkoutBranch)); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "checkout", refToMergeInto)); err != nil {
		return
	}

	if originalHead, err = gitGetHeadSha(slaveRepository); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "merge", "FETCH_HEAD", "-m", fmt.Sprintf("Merging in %s", refSha))); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "push", fmt.Sprintf("HEAD:%s", refToMergeInto))); err != nil {
		return
	}

	return
}

func gitGetHeadSha(repository *Repository) (headSha string, err error) {
	showCommand := repository.Command(nil, "show", "HEAD")
	if err = RunCommand(showCommand); err != nil {
		return
	}

	//TODO(akostov) This line sucks. Does anyone know of a better way to get a bufio.Reader from a string?
	headDataReader := bufio.NewReader(strings.NewReader(showCommand.Stdout.String()))
	shaLine, err := headDataReader.ReadString('\n')
	if err != nil {
		return
	}

	headSha = strings.TrimPrefix(shaLine, "commit ")
	if headSha == shaLine {
		err = fmt.Errorf("git show HEAD output data for repository at %s was not formatted as expected.", repository.path)
		return
	}
	return
}

func gitResetRepositoryHead(repository *Repository, refToReset, originalHead string) error {
	return RunCommand(repository.Command(nil, "push", "origin", fmt.Sprintf("%s:%s", originalHead, refToReset), "--force"))
}

func gitUpdateBranchFromForwardUrl(repository *Repository, remoteUri, refToUpdate string) (headSha string, err error) {
	remoteBranch := fmt.Sprintf("origin/%s", refToUpdate)

	if err = gitFetchWithPrivateKey(repository, remoteUri, refToUpdate); err != nil {
		return
	}

	if err = RunCommand(repository.Command(nil, "checkout", "FETCH_HEAD")); err != nil {
		return
	}

	if headSha, err = gitGetHeadSha(repository); err != nil {
		return
	}

	if err = RunCommand(repository.Command(nil, "branch", "-f", refToUpdate, remoteBranch)); err != nil {
		return
	}

	if err = RunCommand(repository.Command(nil, "checkout", refToUpdate)); err != nil {
		return
	}

	return
}

func gitUpdateFromForwardUrl(slaveRepository *Repository, remoteUri, refToMergeInto, originalHead string) (err error) {
	defer func() {
		if err != nil {
			gitResetRepositoryHead(slaveRepository, refToMergeInto, originalHead)
		} else {
			RunCommand(slaveRepository.Command(nil, "reset", "--hard"))
		}
	}()

	refSha, err := gitUpdateBranchFromForwardUrl(slaveRepository, remoteUri, refToMergeInto)
	if err != nil {
		return
	}

	remoteBranch := fmt.Sprintf("origin/%s", refToMergeInto)

	if err = RunCommand(slaveRepository.Command(nil, "checkout", remoteBranch)); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "branch", "-f", refToMergeInto, remoteBranch)); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "checkout", refToMergeInto)); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "merge", "FETCH_HEAD", "-m", fmt.Sprintf("Merging in %s", refSha))); err != nil {
		return
	}

	if err = RunCommand(slaveRepository.Command(nil, "push", "origin", fmt.Sprintf("HEAD:%s", refToMergeInto))); err != nil {
		return
	}

	return
}

func gitPushMergeRetry(bareRepository, slaveRepository *Repository, remoteUri, refToMergeInto, originalHead string) (err error) {
	i := 0

	for {
		err = gitPushWithPrivateKey(bareRepository, remoteUri, fmt.Sprintf("%s:%s", refToMergeInto, refToMergeInto))

		//TODO(akostov) More precise error catching
		if err != nil && i < pushMergeRetries {
			i++
			time.Sleep(retryTimeout)
			gitUpdateFromForwardUrl(slaveRepository, remoteUri, refToMergeInto, originalHead)
		} else if err != nil {
			gitResetRepositoryHead(slaveRepository, refToMergeInto, originalHead)
		}
	}
	return
}

func gitGetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return
	}

	storedRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return
	}

	command := storedRepository.Command(nil, "show", ref)
	if err = RunCommand(command); err != nil {
		err = NoSuchCommitInRepositoryError{fmt.Sprintf(fmt.Sprintf("The repository %v does not contain commit %s", repository, ref))}
		return
	}

	commitDataReader := bufio.NewReader(strings.NewReader(command.Stdout.String()))

	shaLine, err := commitDataReader.ReadString('\n')
	if err != nil {
		return
	}

	if !strings.HasPrefix(shaLine, "commit ") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	authorLine, err := commitDataReader.ReadString('\n')

	author := strings.TrimPrefix(authorLine, "Author: ")
	if author == authorLine {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	authorSplit := strings.Split(author, " <")

	username = strings.Trim(authorSplit[0], " ")
	email = strings.Trim(authorSplit[1], "> ")

	dateLine, err := commitDataReader.ReadString('\n')

	if !strings.HasPrefix(dateLine, "Date:") {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	blankLine, err := commitDataReader.ReadString('\n')

	if blankLine != "" {
		err = fmt.Errorf("git show %s output data for repository at %v was not formatted as expected.", ref, repository)
		return
	}

	messageLine, err := commitDataReader.ReadString('\n')

	message = strings.Trim(messageLine, " ")

	return
}

func gitGetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return
	}

	storedRepository, err := Open(repository.VcsType, path)
	if err != nil {
		return
	}

	//TODO(akostob + bbland) Discuss whether we should just do it here or have it be explicit.
	if err = gitStorePending(storedRepository, ref, repository.RemoteUri); err != nil {
		return
	}

	// TODO(akostov) discuss getting rid of .koality.yml file
	command := storedRepository.Command(nil, "show", fmt.Sprintf("%s:koality.yml", ref))
	if err = RunCommand(command); err != nil {
		return
	}

	yamlFile = command.Stdout.String()
	return
}
