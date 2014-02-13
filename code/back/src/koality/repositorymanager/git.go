package repositorymanager

import (
	"fmt"
	"koality/resources"
	"koality/shell"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	pushMergeRetries = 4
	retryTimeout     = time.Second
)

type gitSubRepository struct {
	path                string
	resourcesConnection *resources.Connection
}

func (repository *gitSubRepository) getVcsBaseCommand() string {
	return "git"
}

func (repository *gitSubRepository) getPath() string {
	return repository.path
}

type gitRepository struct {
	bare  *gitSubRepository
	slave *gitSubRepository

	name      string
	remoteUri string
}

var ensureGitInstalledCommand = shell.And(
	shell.Or(
		shell.Silent("which ssh"),
		shell.Advertised(shell.Sudo("apt-get install -y ssh-client")),
	),
	shell.Or(
		shell.Silent("which git"),
		shell.Advertised(shell.Sudo("apt-get install -y git-core")),
	),
)

func (repositoryManager *repositoryManager) openGitRepository(repository *resources.Repository) *gitRepository {
	path := repositoryManager.ToPath(repository)
	return &gitRepository{&gitSubRepository{path, repositoryManager.resourcesConnection}, &gitSubRepository{path + ".slave", repositoryManager.resourcesConnection}, repository.Name, repository.RemoteUri}
}

func (repository *gitSubRepository) getSshEnvironment() []string {
	keyPair, err := repository.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		return []string{}
	}

	return []string{
		fmt.Sprintf("GIT_SSH=%s", defaultSshScript),
		fmt.Sprintf("SSH_PRIVATE_KEY=%s", keyPair.PrivateKey),
		fmt.Sprintf("SSH_TIMEOUT=%s", defaultTimeout),
	}
}

func (repository *gitSubRepository) pruneExcessBranches(remoteUri string) (err error) {
	env := repository.getSshEnvironment()

	if err := RunCommand(Command(repository, env, "remote", "prune", remoteUri)); err != nil {
		return err
	}

	excessBranches := make(map[string]bool)

	command := Command(repository, nil, "branch")
	if err := RunCommand(command); err != nil {
		return err
	}

	for {
		branchLine, err := command.Stdout.ReadString('\n')
		if err != nil {
			break
		}

		if !strings.HasPrefix(strings.Trim(strings.TrimSpace(branchLine), " "), "*") {
			excessBranches[strings.TrimSpace(branchLine)] = true
		}
	}

	command = Command(repository, nil, "ls-remote", "-h", remoteUri)
	if err := RunCommand(command); err != nil {
		return err
	}

	for {
		branchLine, err := command.Stdout.ReadString('\n')
		if err != nil {
			break
		}

		branch := strings.TrimPrefix(strings.TrimSpace(strings.Split(branchLine, "\t")[1]), "refs/heads/")

		if _, ok := excessBranches[branch]; ok {
			excessBranches[branch] = false
		}
	}

	for branch, excess := range excessBranches {
		if excess {
			if err := RunCommand(Command(repository, nil, "branch", "-d", branch)); err != nil {
				return err
			}
		}
	}

	return
}

func (repository *gitSubRepository) fetchWithPrivateKey(remoteUri string, args ...string) (err error) {
	env := repository.getSshEnvironment()

	if err := repository.pruneExcessBranches(remoteUri); err != nil {
		return err
	}

	if err := RunCommand(Command(repository, env, "fetch", append([]string{remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func (repository *gitSubRepository) pushWithPrivateKey(remoteUri string, args ...string) (err error) {
	env := repository.getSshEnvironment()

	if err := RunCommand(Command(repository, env, "push", append([]string{remoteUri}, args...)...)); err != nil {
		return err
	}

	return
}

func (repository *gitSubRepository) fetchAllRefsExceptFor(remoteUri string) (err error) {
	if err = repository.fetchWithPrivateKey(remoteUri, "+refs/*:refs/*"); err != nil {
		return
	}

	command := Command(repository, nil, "for-each-ref", fmt.Sprintf("--format=%s", shell.Quote("%(refname)")), "refs/for/", "refs/force/")
	if err = RunCommand(command); err != nil {
		return
	}

	for {
		refLine, err := command.Stdout.ReadString('\n')
		if err != nil {
			break
		}

		ref := strings.Trim(strings.TrimSpace(refLine), "'")
		command := Command(repository, nil, "update-ref", "-d", ref)
		if err = RunCommand(command); err != nil {
			return err
		}

		if err = repository.pushWithPrivateKey(remoteUri, fmt.Sprintf(":%s", ref)); err != nil {
			return err
		}
	}

	return
}

func (repository *gitRepository) storePending(ref, remoteUri string, args ...string) (err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return
	}

	if err = repository.bare.fetchAllRefsExceptFor(remoteUri); err != nil {
		return
	}

	if err = RunCommand(Command(repository.bare, nil, "show", ref)); err != nil {
		return NoSuchCommitInRepositoryError{fmt.Sprintf("The repository at %s does not contain commit %s", repository.bare.path, ref)}
	}

	if err = repository.bare.pushWithPrivateKey(remoteUri, fmt.Sprintf("%s:%s", ref, GitHiddenRef(ref))); err != nil {
		return
	}

	return
}

func (repository *gitRepository) createRepository() (err error) {
	if _, err = os.Stat(repository.bare.path); !os.IsNotExist(err) {
		return RepositoryAlreadyExistsInStoreError{fmt.Sprintf("The repository at %s already exists in the repository store.", repository.bare.path)}
	}

	if err = os.MkdirAll(repository.bare.path, 0700); err != nil {
		return
	}

	if err := RunCommand(Command(repository.bare, nil, "init", "--bare")); err != nil {
		return err
	}

	if err = repository.bare.fetchAllRefsExceptFor(repository.remoteUri); err != nil {
		return
	}

	if err = RunCommand(Command(repository.bare, nil, "clone", repository.bare.path, repository.slave.path)); err != nil {
		return err
	}

	return
}

func (repository *gitRepository) deleteRepository() (err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return err
	}

	err = os.RemoveAll(repository.slave.path)
	if err != nil {
		return err
	}

	return os.RemoveAll(repository.bare.path)
}

func (repository *gitRepository) mergeChangeset(headRef, baseRef, refToMergeInto string) (err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return
	}

	if strings.HasPrefix(refToMergeInto, "refs/heads/") {
		refToMergeInto = refToMergeInto[len("refs/heads/"):]
	} else {
		return fmt.Errorf("Unable to merge into ref: %s", refToMergeInto)
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

func (repository *gitRepository) mergeRefs(refToMerge, refToMergeInto string) (originalHead string, err error) {
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

	if remoteBranchExists, err := regexp.MatchString(fmt.Sprintf("\\s+%s$", remoteBranch), branches); remoteBranchExists && err != nil {
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

	refSha, err := repository.slave.getTopShaForSubrepository("HEAD")
	if err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "branch", "-f", refToMergeInto, checkoutBranch)); err != nil {
		return
	}

	if err = RunCommand(Command(repository.slave, nil, "checkout", refToMergeInto)); err != nil {
		return
	}

	if originalHead, err = repository.slave.getTopShaForSubrepository("HEAD"); err != nil {
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

func (repository *gitSubRepository) getTopShaForSubrepository(ref string) (topSha string, err error) {
	showCommand := Command(repository, nil, "show", ref)
	if err = RunCommand(showCommand); err != nil {
		return
	}

	shaLine, err := showCommand.Stdout.ReadString('\n')
	if err != nil {
		return
	}

	if !strings.HasPrefix(shaLine, "commit ") {
		err = fmt.Errorf("git show %s output data for repository at %s was not formatted as expected.", ref, repository.path)
		return
	}

	topSha = strings.TrimSpace(strings.TrimPrefix(shaLine, "commit "))
	return
}

func (repository *gitRepository) getTopSha(ref string) (topSha string, err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return
	}

	if err = repository.bare.fetchAllRefsExceptFor(repository.remoteUri); err != nil {
		return
	}

	return repository.slave.getTopShaForSubrepository(ref)
}

func (repository *gitSubRepository) resetRepositoryHead(refToReset, originalHead string) error {
	return RunCommand(Command(repository, nil, "push", "origin", fmt.Sprintf("%s:%s", originalHead, refToReset), "--force"))
}

func (repository *gitSubRepository) updateBranchFromForwardUrl(remoteUri, refToUpdate string) (headSha string, err error) {
	remoteBranch := fmt.Sprintf("origin/%s", refToUpdate)

	if err = repository.fetchWithPrivateKey(remoteUri, refToUpdate); err != nil {
		return
	}

	if err = RunCommand(Command(repository, nil, "checkout", "FETCH_HEAD")); err != nil {
		return
	}

	if headSha, err = repository.getTopShaForSubrepository("HEAD"); err != nil {
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

func (repository *gitRepository) updateFromForwardUrl(remoteUri, refToMergeInto, originalHead string) (err error) {
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

func (repository *gitRepository) pushMergeRetry(remoteUri, refToMergeInto, originalHead string) (err error) {
	for pushAttempt := 0; pushAttempt < pushMergeRetries; pushAttempt++ {
		err = repository.bare.pushWithPrivateKey(remoteUri, fmt.Sprintf("%s:%s", refToMergeInto, refToMergeInto))
		// TODO (aksotov): more precise error checking
		if err != nil && pushAttempt < pushMergeRetries-1 {
			time.Sleep(retryTimeout)
			repository.updateFromForwardUrl(remoteUri, refToMergeInto, originalHead)
		}
	}
	if err != nil {
		repository.slave.resetRepositoryHead(refToMergeInto, originalHead)
	}
	return
}

func (repository *gitRepository) getCommitAttributes(ref string) (headSha, message, username, email string, err error) {
	if err = checkRepositoryExists(repository.bare.path); err != nil {
		return
	}

	if err = repository.bare.fetchAllRefsExceptFor(repository.remoteUri); err != nil {
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
		err = fmt.Errorf("git show %s output data for repository at %s was not formatted as expected.", ref, repository)
		return
	}

	headSha = strings.TrimSpace(strings.TrimPrefix(shaLine, "commit "))

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

func (repository *gitRepository) getYamlFile(ref string) (yamlFile string, err error) {
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

func (repository *gitRepository) getCloneCommand() shell.Command {
	return shell.And(
		ensureGitInstalledCommand,
		shell.Or(
			shell.Not(shell.Test(shell.Commandf("-e %s", repository.name))),
			shell.Advertised(shell.Commandf("rm -rf %s", repository.name)),
		),
		shell.Advertised(shell.Commandf("git clone %s %s", repository.remoteUri, repository.name)),
	)
}

func (repository *gitRepository) getCheckoutCommand(ref, baseRef string) shell.Command {
	commands := []shell.Command{
		ensureGitInstalledCommand,
		shell.Or(
			shell.Test(shell.Commandf("-d %s", repository.name)),
			shell.Advertised(shell.Commandf("git init %s", repository.name)),
		),
		shell.Advertised(shell.Commandf("cd %s", repository.name)),
		shell.Advertised(shell.Commandf("git fetch %s %s -n --depth 1", repository.remoteUri, GitHiddenRef(ref))),
		shell.Advertised(shell.Commandf("git checkout --force %s", ref)),
	}
	if baseRef != ref {
		commands = append(commands, shell.If(
			shell.Silent(shell.Commandf("git show-ref %s", baseRef)),
			shell.Advertised(shell.Commandf("git merge --no-edit %s", baseRef)),
		))
	}
	if strings.HasPrefix(baseRef, "refs/heads/") {
		branch := baseRef[len("refs/heads/"):]
		commands = append(commands, shell.Advertised(shell.Commandf("git checkout -B %s", branch)))
	}
	return shell.And(commands...)
}
