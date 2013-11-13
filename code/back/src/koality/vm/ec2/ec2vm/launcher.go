package ec2vm

import (
	"bytes"
	"fmt"
	"github.com/crowdmob/goamz/ec2"
	"io"
	"io/ioutil"
	"koality/shell"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"mime/multipart"
	"net/textproto"
	"os"
	"strings"
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
	sshAttemptTimeout := 3 * time.Minute
	sshAttemptTimeoutChan := time.After(sshAttemptTimeout)
	for {
		select {
		case <-sshAttemptTimeoutChan:
			return nil, fmt.Errorf("Failed to ssh into the instance after %s", sshAttemptTimeout.String())
		default:
			ec2Vm, err := New(&instance, launcher.ec2Broker, username)
			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}
			sshAttempt, err := ec2Vm.MakeExecutable(shell.Command("true"))
			if err == nil {
				err = sshAttempt.Run()
				if err == nil {
					return ec2Vm, nil
				}
			}
			time.Sleep(3 * time.Second)
		}
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

	buffer := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(buffer)
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\nMIME-Version: 1.0\n\n", multipartWriter.Boundary()))

	err = writeCloudInitMimePart(multipartWriter, fmt.Sprintf("#!/bin/sh\n%s", configureUserCommand), "koality-data")
	if err != nil {
		panic(err)
	}
	multipartWriter.Close()

	return buffer.Bytes()
}

func writeCloudInitMimePart(multipartWriter *multipart.Writer, contents, name string) error {
	mimeHeader := make(textproto.MIMEHeader)
	mimeHeader.Set("Content-Type", cloudInitMimeType(contents))
	mimeHeader.Set("Content-Disposition", fmt.Sprintf("form-data; name=\"%s\"", name))
	partWriter, err := multipartWriter.CreatePart(mimeHeader)
	if err != nil {
		return err
	}
	io.WriteString(partWriter, contents)
	return nil
}

func cloudInitMimeType(contents string) string {
	startsWithMapping := map[string]string{
		"#include":              "text/x-include-url",
		"#!":                    "text/x-shellscript",
		"#cloud-boothook":       "text/cloud-boothook",
		"#cloud-config":         "text/cloud-config",
		"#cloud-config-archive": "text/cloud-config-archive",
		"#upstart-job":          "text/upstart-job",
		"#part-handler":         "text/part-handler",
	}
	for prefix, contentType := range startsWithMapping {
		if strings.HasPrefix(contents, prefix) {
			return contentType
		}
	}
	return "text/plain"
}
