package ec2vm

import (
	"fmt"
	"github.com/crowdmob/goamz/ec2"
	"koality/shell"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"time"
)

type EC2VirtualMachineLauncher struct {
	ec2Broker *ec2broker.EC2Broker
}

func NewLauncher(ec2Broker *ec2broker.EC2Broker) *EC2VirtualMachineLauncher {
	return &EC2VirtualMachineLauncher{ec2Broker}
}

func (launcher *EC2VirtualMachineLauncher) LaunchVirtualMachine() vm.VirtualMachine {
	username := launcher.getUsername()
	runOptions := ec2.RunInstancesOptions{
		ImageId:        launcher.getImageId(),
		InstanceType:   launcher.getInstanceType(),
		SecurityGroups: launcher.getSecurityGroups(),
		UserData:       launcher.getUserData(username),
	}
	runResponse, err := launcher.ec2Broker.EC2().RunInstances(&runOptions)
	if err != nil {
		panic(err)
	}
	// TODO: more validation
	instance := runResponse.Instances[0]
	for instance.IPAddress == "" {
		time.Sleep(time.Duration(5) * time.Second)
		instances := launcher.ec2Broker.Instances()
		for _, inst := range instances {
			if inst.InstanceId == instance.InstanceId {
				instance = inst
				break
			}
		}
	}

	ec2Vm := New(&instance, launcher.ec2Broker, username)
	for {
		sshAttempt := ec2Vm.MakeExecutable(shell.Command("true"))
		err := sshAttempt.Run()
		if err == nil {
			break
		}
		time.Sleep(time.Duration(3) * time.Second)
		ec2Vm = New(&instance, launcher.ec2Broker, username)
	}
	return ec2Vm
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
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCYn9+XcfZjSscco4g8BNXY0Ap9Zdrc+fkigorpoMNhaoo4QWxHctsQ0f8VWQ39+Vn6nTGwWipSmGnPKN2kL7dFv3PT1nEC7+iT2skQELG0YC8j84CaMFY2NuDsSvJSE0O9Xm7Bq8O1bvLyrL8qRNrD9/CQSl1viUHr00P3wMwSFYq2BkfuUC3rVQttV6z1TQjMJjeg+YeKCZkzdpHx6a6Ay/M1fKg5ZXDGcJDKg5s783SOOSyvpzhvIWbuTTBcHmrgW/xPiqPAP5LmRrJkHZd/V95TPNIZ6duFEasOykjU8h9h7He4Cim4IfnNVvKwU62Xs9Mp5FESm9ozr3PP1bcj"

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