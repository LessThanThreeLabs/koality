package ec2vm

import (
	"errors"
	"fmt"
	"github.com/crowdmob/goamz/ec2"
	"io"
	"koality/shell"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
)

type EC2VirtualMachine struct {
	sshExecutableMaker *vm.SshExecutableMaker
	fileCopier         shell.FileCopier
	patcher            vm.Patcher
	instance           *ec2.Instance
	ec2Broker          *ec2broker.EC2Broker
}

func New(instance *ec2.Instance, broker *ec2broker.EC2Broker, username string) (*EC2VirtualMachine, error) {
	sshConfig := vm.SshConfig{
		Username: username,
		Hostname: instance.IPAddress,
		Port:     22,
		Options: map[string]string{
			"LogLevel":              "error",
			"StrictHostKeyChecking": "no",
			"UserKnownHostsFile":    "/dev/null",
			"ServerAliveInterval":   "20",
		},
	}
	sshExecutableMaker, err := vm.NewSshExecutableMaker(sshConfig)
	if err != nil {
		return nil, err
	}
	scper, err := vm.NewScper(vm.ScpConfig(sshConfig))
	if err != nil {
		return nil, err
	}
	fileCopier := &vm.ScpFileCopier{scper}
	patcher := vm.NewPatcher(fileCopier, sshExecutableMaker)
	ec2Vm := EC2VirtualMachine{
		sshExecutableMaker: sshExecutableMaker,
		fileCopier:         fileCopier,
		patcher:            patcher,
		instance:           instance,
		ec2Broker:          broker,
	}
	return &ec2Vm, nil
}

func (ec2Vm *EC2VirtualMachine) MakeExecutable(command shell.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer) (shell.Executable, error) {
	return ec2Vm.sshExecutableMaker.MakeExecutable(command, stdin, stdout, stderr)
}

func (ec2Vm *EC2VirtualMachine) ProvisionCommand() shell.Command {
	panic("Not implemented")
	return shell.Command("false")
}

func (ec2Vm *EC2VirtualMachine) Patch(patchConfig *vm.PatchConfig) (shell.Executable, error) {
	return ec2Vm.patcher.Patch(patchConfig)
}

func (ec2Vm *EC2VirtualMachine) FileCopy(sourceFilePath, destFilePath string) (shell.Executable, error) {
	return ec2Vm.fileCopier.FileCopy(sourceFilePath, destFilePath)
}

func (ec2Vm *EC2VirtualMachine) Terminate() error {
	terminateResp, err := ec2Vm.ec2Broker.EC2().TerminateInstances([]string{ec2Vm.instance.InstanceId})
	if err != nil {
		return err
	}
	if len(terminateResp.StateChanges) != 1 {
		return errors.New(fmt.Sprintf("Expected one state change, found %v instead", terminateResp.StateChanges))
	}
	stateChange := terminateResp.StateChanges[0]
	if stateChange.CurrentState.Name != "shutting-down" && stateChange.CurrentState.Name != "terminated" {
		return errors.New(fmt.Sprintf("Expected new state to be \"shutting-down\" or \"terminated\", was %q instead", stateChange.CurrentState.Name))
	}
	return nil
}
