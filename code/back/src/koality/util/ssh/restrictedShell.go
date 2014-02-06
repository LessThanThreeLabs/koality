package ssh

import (
	"koality/internalapi"
	"koality/resources"
	"koality/vm"
	"net/rpc"
	"strings"
)

type shell interface {
	GetCommand() (vm.Command, error)
	GetClient() *rpc.Client
}

func HandleSSHCommand(userId uint64, origCommand string) error {
	command := strings.Split(origCommand, " ")
	client, err := rpc.Dial("unix", internalapi.RpcSocket)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(command) == 0 {
		user := new(resources.User)
		if err := client.Call("UserInfoReader.GetUser", &userId, user); err != nil {
			return err
		}

		return NoShellAccessError{user.Email}
	}

	var shellCommand vm.Command
	switch command[0] {
	case "true":
		// true
		shellCommand = vm.Command{Argv: []string{"true"}}
	case "ssh":
		// ssh poolId instanceId
		shell := &restrictedSSHForwardingShell{userId, command, client}
		shellCommand, err = shell.GetCommand()
	default:
		// git-receive-pack localuri.git
		// git-upload-pack localuri.git
		shell := &restrictedGitShell{userId, command, client}
		shellCommand, err = shell.GetCommand()
	}
	if err != nil {
		return err
	}

	return shellCommand.Exec()
}
