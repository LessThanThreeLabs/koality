package ec2vm

import (
	"errors"
	"fmt"
	"github.com/crowdmob/goamz/ec2"
	"koality/shell"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
)

type EC2VirtualMachine struct {
	sshExecutableMaker *vm.SshExecutableMaker
	scper              vm.Scper
	patcher            *vm.ScpPatcher
	instance           *ec2.Instance
	ec2Broker          *ec2broker.EC2Broker
}

func New(instance *ec2.Instance, broker *ec2broker.EC2Broker, username string) *EC2VirtualMachine {
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
	sshExecutableMaker := vm.NewSshExecutableMaker(sshConfig)
	scper := vm.NewScper(vm.ScpConfig(sshConfig))
	patcher := &vm.ScpPatcher{
		Scper:           scper,
		ExecutableMaker: sshExecutableMaker,
	}
	return &EC2VirtualMachine{
		sshExecutableMaker: sshExecutableMaker,
		scper:              scper,
		patcher:            patcher,
		instance:           instance,
		ec2Broker:          broker,
	}
}

func (ec2Vm *EC2VirtualMachine) MakeExecutable(command shell.Command) shell.Executable {
	return ec2Vm.sshExecutableMaker.MakeExecutable(command)
}

func (ec2Vm *EC2VirtualMachine) ProvisionCommand() shell.Command {
	panic("Not implemented")
	return shell.Command("false")
}

func (ec2Vm *EC2VirtualMachine) Patch(patchConfig *vm.PatchConfig) (shell.Executable, error) {
	return ec2Vm.patcher.Patch(patchConfig)
}

func (ec2Vm *EC2VirtualMachine) FileCopy(sourceFilePath, destFilePath string) shell.Executable {
	return ec2Vm.scper.Scp(sourceFilePath, destFilePath, false)
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
	if stateChange.CurrentState.Name != "terminated" {
		return errors.New(fmt.Sprintf("Expected new state to be \"terminated\", was %q instead", stateChange.CurrentState.Name))
	}
	return nil
}
