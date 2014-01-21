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

func (shell *restrictedSSHForwardingShell) DoCommand() (err error) {
	if len(shell.command) != 4 {
		err = InvalidCommandError{strings.Join(shell.command, " ")}
		return
	}

	vmInstanceId := shell.command[1]
	poolId, err := strconv.ParseUint(shell.command[2], 10, 64)
	if err != nil {
		return
	}

	var vm vm.VirtualMachine

	if err = verifyUserExists(shell.client, shell.userId, nil); err != nil {
		return
	}

	args := internalapi.VmReaderArg{vmInstanceId, poolId}
	if err = shell.client.Call("VmReader.GetVmFromId", &args, &vm); err != nil {
		return
	}

	exec, err := vm.MakeExecutable("bash", os.Stdin, os.Stdout, os.Stderr, nil)
	if err != nil {
		return
	}

	err = exec.Run()
	return
}
