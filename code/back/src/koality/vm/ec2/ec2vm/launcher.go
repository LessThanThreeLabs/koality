package ec2vm

import (
	"fmt"
	"github.com/crowdmob/goamz/ec2"
	"koality/shell"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"os"
	"time"
)

type EC2VirtualMachineLauncher struct {
	ec2Broker *ec2broker.EC2Broker
}

func NewLauncher(ec2Broker *ec2broker.EC2Broker) *EC2VirtualMachineLauncher {
	return &EC2VirtualMachineLauncher{ec2Broker}
}

func (launcher *EC2VirtualMachineLauncher) LaunchVirtualMachine() (vm.VirtualMachine, error) {
	username := launcher.getUsername()
	runOptions := ec2.RunInstancesOptions{
		ImageId:        launcher.getImageId(),
		InstanceType:   launcher.getInstanceType(),
		SecurityGroups: launcher.getSecurityGroups(),
		UserData:       launcher.getUserData(username),
	}
	runResponse, err := launcher.ec2Broker.EC2().RunInstances(&runOptions)
	if err != nil {
		return nil, err
	}
	// TODO: more validation
	instance := runResponse.Instances[0]
	for instance.IPAddress == "" {
		time.Sleep(5 * time.Second)
		instances := launcher.ec2Broker.Instances()
		for _, inst := range instances {
			if inst.InstanceId == instance.InstanceId {
				instance = inst
				break
			}
		}
	}

	for {
		ec2Vm, err := New(&instance, launcher.ec2Broker, username)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		sshAttempt, err := ec2Vm.MakeExecutable(shell.Command("true"))
		if err == nil {
			err = sshAttempt.Run()
			if err == nil {
				return ec2Vm
			}
		}
		time.Sleep(3 * time.Second)
	}
}

// TODO: make all this stuff dynamic
func (launcher *EC2VirtualMachineLauncher) getUsername() string {
	return "lt3"
}

func (launcher *EC2VirtualMachineLauncher) getImageId() string {
	return "ami-2cba241c"
}

func (launcher *EC2VirtualMachineLauncher) getInstanceType() string {
	return "m1.small"
}

func (launcher *EC2VirtualMachineLauncher) getSecurityGroups() []ec2.SecurityGroup {
	return []ec2.SecurityGroup{
		ec2.SecurityGroup{
			Name: "koality_verification",
		},
	}
}

func (launcher *EC2VirtualMachineLauncher) getUserData(username string) []byte {
	keyBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/.ssh/id_rsa.pub", os.Getenv("HOME")))
	if err != nil {
		panic(err)
	}
	publicKey := string(keyBytes)

	configureUserCommand := shell.Chain(
		shell.Command(fmt.Sprintf("useradd --create-home -s /bin/bash %s", username)),
		shell.Command(fmt.Sprintf("mkdir ~%s/.ssh", username)),
		shell.Append(shell.Command(fmt.Sprintf("echo %s", shell.Quote(publicKey))), shell.Command(fmt.Sprintf("~%s/.ssh/authorized_keys", username)), false),
		shell.Command(fmt.Sprintf("chown -R %s:%s ~%s/.ssh", username, username, username)),
		shell.Or(
			shell.Command("grep '#includedir /etc/sudoers.d' /etc/sudoers"),
			shell.Append(shell.Command("echo #includedir /etc/sudoers.d"), shell.Command("/etc/sudoers"), false),
		),
		shell.Command("mkdir /etc/sudoers.d"),
		shell.Redirect(shell.Command(fmt.Sprintf("echo 'Defaults !requiretty\n%s ALL=(ALL) NOPASSWD: ALL'", username)), shell.Command(fmt.Sprintf("/etc/sudoers.d/koality-%s", username)), false),
		shell.Command(fmt.Sprintf("chmod 0440 /etc/sudoers.d/koality-%s", username)),
	)
	return ([]byte)(fmt.Sprintf("#!/bin/sh\n%s", configureUserCommand))
}
