package ssh

import (
	"fmt"
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
		shellCommand = vm.Command{Argv: []string{"true"}}
	case "ssh":
		shell := &restrictedSSHForwardingShell{userId, command, client}
		shellCommand, err = shell.GetCommand()
	default:
		shell := &restrictedGitShell{userId, command, client}
		shellCommand, err = shell.GetCommand()
	}
	if err != nil {
		return err
	}

        // TODO(dhuang) remove this later
	fmt.Println("command:\n", shellCommand)
	return shellCommand.Exec()
}
