package ssh

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"koality/internalapi"
	"koality/repositorymanager"
	"koality/resources"
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
		"git-receive-pack": 3,
		"git-upload-pack":  3,
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

	// command[1] should always be localUri regardless of command
	localUri := strings.Trim(shell.command[1], "'")
	if !strings.HasSuffix(localUri, ".git") {
		localUri += ".git"
	}
	if strings.Contains(localUri, "..") {
		err = MalformedCommandError{localUri}
		return
	}

	var repositoryInfo internalapi.RepositoryInfo
	repoCall := shell.client.Go("RepositoryReader.GetRepoFromLocalUri",
		&localUri, &repositoryInfo, nil)
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

		repositoryManager := repositorymanager.New(repositoryInfo.RepositoriesPath)
		repoPath := repositoryManager.ToPath(repository)
		command.Argv = []string{"jgit", "receive-pack", repoPath, string(shell.userId)}
	case "git-upload-pack":
		var privateKey string
		var emptyInput interface{}
		if err = shell.client.Call("UserInfoReader.GetRepoPrivateKey", &emptyInput, &privateKey); err != nil {
			return
		}
		hash := sha1.New()
		_, err = io.WriteString(hash, privateKey)
		if err != nil {
			return
		}

		privateKeyPath := path.Join("/tmp", hex.EncodeToString(hash.Sum(nil)))
		err = ioutil.WriteFile(privateKeyPath, []byte(privateKey), 0600)
		if err != nil {
			return
		}

		<-repoCall.Done
		if err = checkRepoCallError(repoCall); err != nil {
			return
		}

		remoteUriParts := strings.Split(repository.RemoteUri, ":")
		uri := remoteUriParts[0]
		path := remoteUriParts[1]
		command.Argv = []string{"ssh", "-oStrictHostKeyChecking=no",
			"-i", privateKeyPath, uri, "git-upload-pack " + path}
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
