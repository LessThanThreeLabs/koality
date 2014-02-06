package ssh

import (
	"koality/internalapi"
	"koality/vm"
	"net/rpc"
	"os"
	"strconv"
	"strings"
)

type restrictedSSHForwardingShell struct {
	userId  uint64
	command []string
	client  *rpc.Client
}

func (shell *restrictedSSHForwardingShell) GetCommand() (command vm.Command, err error) {
	if len(shell.command) != 3 {
		err = InvalidCommandError{strings.Join(shell.command, " ")}
		return
	}

	poolId, err := strconv.ParseUint(shell.command[1], 10, 64)
	if err != nil {
		return
	}
	vmInstanceId := shell.command[2]

	args := internalapi.VmReaderArg{vmInstanceId, poolId}
	err = shell.client.Call("VmReader.GetShellCommandFromId", &args, &command)
	command.Envv = append(command.Envv, os.Environ()...)
	return
}
