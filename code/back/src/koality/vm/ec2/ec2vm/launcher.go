package ec2vm

import (
	"bytes"
	"errors"
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
	"strconv"
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
	activeImage, err := launcher.getActiveImage()
	if err != nil {
		return nil, err
	}
	runOptions := ec2.RunInstancesOptions{
		ImageId:        activeImage.Id,
		InstanceType:   launcher.getInstanceType(),
		SecurityGroups: launcher.getSecurityGroups(),
		UserData:       launcher.getUserData(username),
	}
	runResponse, err := launcher.ec2Broker.EC2().RunInstances(&runOptions)
	if err != nil {
		return nil, err
	}
	// TODO (bbland): add more validation
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

// TODO (bbland): make all this stuff dynamic
func (launcher *EC2VirtualMachineLauncher) getUsername() string {
	return "lt3"
}

func (launcher *EC2VirtualMachineLauncher) getActiveImage() (ec2.Image, error) {
	baseImage, err := launcher.getBaseImage()
	if err != nil {
		return ec2.Image{}, err
	}
	snapshots, err := launcher.getSnapshotsForImage(baseImage)
	if err != nil || len(snapshots) == 0 {
		return baseImage, nil
	}
	newestSnapshot := snapshots[0]
	newestSnapshotVersion, _ := launcher.getSnapshotVersion(newestSnapshot)
	for _, snapshot := range snapshots {
		snapshotVersion, err := launcher.getSnapshotVersion(snapshot)
		if err != nil && snapshotVersion > newestSnapshotVersion {
			newestSnapshotVersion = snapshotVersion
			newestSnapshot = snapshot
		}
	}
	return newestSnapshot, nil
}

func (launcher *EC2VirtualMachineLauncher) getSnapshotVersion(snapshot ec2.Image) (int, error) {
	versionIndex := strings.LastIndex(snapshot.Name, "v") + 1
	if versionIndex == 0 {
		return -1, errors.New("Could not find version in snapshot name")
	}
	version, err := strconv.Atoi(snapshot.Name[versionIndex:])
	if err != nil {
		return -1, errors.New("Could not find version in snapshot name")
	}
	return version, nil
}

func (launcher *EC2VirtualMachineLauncher) getBaseImage() (ec2.Image, error) {
	// TODO (bbland): try to use a specified base image instead of just the default
	imageFilter := ec2.NewFilter()
	imageFilter.Add("owner-id", "600991114254") // must be changed if our ec2 info changes
	imageFilter.Add("name", "koality_verification_precise_0.4")
	imageFilter.Add("state", "available")
	imagesResp, err := launcher.ec2Broker.EC2().Images([]string{}, imageFilter)
	if err != nil {
		return ec2.Image{}, err
	}
	return imagesResp.Images[0], nil
}

func (launcher *EC2VirtualMachineLauncher) getSnapshotsForImage(baseImage ec2.Image) ([]ec2.Image, error) {
	imageFilter := ec2.NewFilter()
	imageFilter.Add("owner-alias", "self")
	imageFilter.Add("name", fmt.Sprintf("koality-snapshot-(%s)-v*", baseImage.Name))
	imageFilter.Add("state", "available")
	imagesResp, err := launcher.ec2Broker.EC2().Images([]string{}, imageFilter)
	if err != nil {
		return []ec2.Image{}, err
	}
	return imagesResp.Images, nil
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
	buffer := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(buffer)
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\nMIME-Version: 1.0\n\n", multipartWriter.Boundary()))

	err := writeCloudInitMimePart(multipartWriter, launcher.getDefaultUserData(username), "koality-data")
	if err != nil {
		panic(err)
	}
	multipartWriter.Close()

	return buffer.Bytes()
}

func (launcher *EC2VirtualMachineLauncher) getDefaultUserData(username string) string {
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
	return fmt.Sprintf("#!/bin/sh\n%s", configureUserCommand)
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
