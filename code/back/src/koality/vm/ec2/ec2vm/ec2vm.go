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

type Ec2VirtualMachine struct {
	sshExecutableMaker *vm.SshExecutableMaker
	sshConfig          vm.SshConfig
	fileCopier         shell.FileCopier
	patcher            vm.Patcher
	instance           ec2.Instance
	ec2Cache           *ec2broker.Ec2Cache
}

func New(instance ec2.Instance, cache *ec2broker.Ec2Cache, username, privateKey string) (*Ec2VirtualMachine, error) {
	sshConfig := vm.SshConfig{
		Username:   username,
		Hostname:   instance.IPAddress,
		Port:       22,
		PrivateKey: privateKey,
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
	ec2Vm := Ec2VirtualMachine{
		sshExecutableMaker: sshExecutableMaker,
		sshConfig:          sshConfig,
		fileCopier:         fileCopier,
		patcher:            patcher,
		instance:           instance,
		ec2Cache:           cache,
	}
	return &ec2Vm, nil
}

func (ec2vm Ec2VirtualMachine) Id() string {
	return ec2vm.instance.InstanceId
}

func (ec2vm Ec2VirtualMachine) GetStartShellCommand() vm.Command {
	return vm.Command{
		Argv: ec2vm.sshConfig.SshArgs(""),
		Envv: []string{"PRIVATE_KEY=" + ec2vm.sshConfig.PrivateKey},
	}
}

func (ec2Vm *Ec2VirtualMachine) MakeExecutable(command shell.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (shell.Executable, error) {
	return ec2Vm.sshExecutableMaker.MakeExecutable(command, stdin, stdout, stderr, environment)
}

func (ec2Vm *Ec2VirtualMachine) Patch(patchConfig *vm.PatchConfig) (shell.Executable, error) {
	return ec2Vm.patcher.Patch(patchConfig)
}

func (ec2Vm *Ec2VirtualMachine) FileCopy(sourceFilePath, destFilePath string) (shell.Executable, error) {
	return ec2Vm.fileCopier.FileCopy(sourceFilePath, destFilePath)
}

func (ec2Vm *Ec2VirtualMachine) Terminate() error {
	terminateResp, err := ec2Vm.ec2Cache.EC2.TerminateInstances([]string{ec2Vm.instance.InstanceId})
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

type commandWithEnv struct {
	Argv []string
	Envv []string
}
