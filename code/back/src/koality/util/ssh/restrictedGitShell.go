package ssh

import (
	"fmt"
	"koality/internalapi"
	"koality/repositorymanager"
	"koality/resources"
	"koality/util/pathtranslator"
	"koality/vm"
	"net/rpc"
	"path"
	"strings"
)

type restrictedGitShell struct {
	userId  uint64
	command []string
	client  *rpc.Client
}

var (
	gitCommandArgNums = map[string]int{
		"git-receive-pack": 2,
		"git-upload-pack":  2,
		// "git-show":         4,
	}
	validGitCommands = map[string]bool{
		"git-receive-pack": true,
		"git-upload-pack":  true,
		// "git-show":         true,
	}
)

func (shell *restrictedGitShell) GetCommand() (command vm.Command, err error) {
	if !validGitCommands[shell.command[0]] {
		err = InvalidCommandError{shell.command[0]}
		return
	} else if len(shell.command) != gitCommandArgNums[shell.command[0]] {
		err = InvalidCommandError{strings.Join(shell.command, " ")}
		return
	}

	// command[1] should always be name regardless of command
	name := strings.TrimSuffix(strings.Trim(shell.command[1], "'"), ".git")
	if strings.Contains(name, "..") {
		err = MalformedCommandError{name}
		return
	}

	var repositoryInfo internalapi.RepositoryInfo
	repoCall := shell.client.Go("RepositoryReader.GetRepoByName",
		&name, &repositoryInfo, nil)
	if err != nil {
		return
	}

	repository := &repositoryInfo.Repository
	switch shell.command[0] {
	case "git-receive-pack":
		<-repoCall.Done
		if err = checkRepoCallError(repoCall); err != nil {
			return
		}

		var gitReceivePack string
		gitReceivePack, err = pathtranslator.TranslatePathWithCheckFunc(path.Join("git", "git-receive-pack"), pathtranslator.CheckExecutable)
		if err != nil {
			return
		}

		repositoryManager := repositorymanager.New(repositoryInfo.RepositoriesPath, nil)
		repoPath := repositoryManager.ToPath(repository)
		command.Argv = []string{gitReceivePack, repoPath}
		command.Envv = []string{fmt.Sprintf("KOALITY_USER_ID=%d", shell.userId)}
	case "git-upload-pack":
		var privateKey string
		var emptyInput interface{}
		if err = shell.client.Call("UserInfoReader.GetRepoPrivateKey", &emptyInput, &privateKey); err != nil {
			return
		}

		<-repoCall.Done
		if err = checkRepoCallError(repoCall); err != nil {
			return
		}

		remoteUriParts := strings.Split(repository.RemoteUri, ":")
		uri := remoteUriParts[0]
		path := remoteUriParts[1]
		var sshwrapperPath string
		sshwrapperPath, err = pathtranslator.TranslatePathAndCheckExists(pathtranslator.BinaryPath("sshwrapper"))
		if err != nil {
			return
		}

		command.Argv = []string{sshwrapperPath, uri, "git-upload-pack " + path}
		command.Envv = []string{"PRIVATE_KEY=" + privateKey}
		// case "git-show":
		// 	show_ref_file := shell.command[2]
		// 	command = []string{"sh", "-c",
		// 		fmt.Sprintf("cd %s && git show %s", repoPath, show_ref_file)}
	}
	return
}

func checkRepoCallError(call *rpc.Call) (err error) {
	if call.Error != nil {
		if _, ok := call.Error.(resources.NoSuchRepositoryError); ok {
			err = RepositoryNotFoundError{}
		} else {
			err = call.Error
		}
	}
	return
}
