package ssh

import (
	"koality/internalapi"
	"koality/resources"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type shell interface {
	DoCommand() ([]string, error)
	GetClient() *rpc.Client
}

func HandleSSHCommand(userId uint64, origCommand string) error {
	command := strings.Split(origCommand, " ")
	client, err := rpc.Dial("unix", internalapi.RpcSocket)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(command) > 0 {
		switch command[0] {
		case "true":
			return execArgv([]string{"true"})
		case "ssh":
			shell := &restrictedSSHForwardingShell{userId, command, client}
			return shell.DoCommand()
		default:
			shell := &restrictedGitShell{userId, command, client}
			return shell.DoCommand()
		}
	} else {
		user, err := lookupUser(client, userId)
		if err != nil {
			return err
		}

		return NoShellAccessError{user.Email}
	}
}

func execArgv(argv []string) error {
	execPath, err := exec.LookPath(argv[0])
	if err != nil {
		return err
	}

	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return err
	}

	return syscall.Exec(absExecPath, argv, os.Environ())
}

func lookupUser(client *rpc.Client, userId uint64) (*resources.User, error) {
	user := new(resources.User)
	if err := client.Call("UserInfoReader.GetUser", &userId, user); err != nil {
		return nil, err
	}

	return user, nil
}

func verifyUserExists(client *rpc.Client, userId uint64, command []string) error {
	user, err := lookupUser(client, userId)
	if err != nil {
		return err
	}

	if user == nil {
		return InvalidPermissionsError{userId, strings.Join(command, " "), 0}
	} else {
		return nil
	}
}

func handleGitCommand(command []string) error {
	return nil
}
