package ec2vm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	"io"
	"io/ioutil"
	"koality/resources"
	"koality/shell"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"mime/multipart"
	"net"
	"net/textproto"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type Ec2VirtualMachineLauncher struct {
	ec2Broker *ec2broker.Ec2Broker
	Ec2Pool   *resources.Ec2Pool
}

func NewLauncher(ec2Broker *ec2broker.Ec2Broker, ec2Pool *resources.Ec2Pool) *Ec2VirtualMachineLauncher {
	return &Ec2VirtualMachineLauncher{ec2Broker, ec2Pool}
}

func (launcher *Ec2VirtualMachineLauncher) ec2Cache() *ec2broker.Ec2Cache {
	auth, err := aws.GetAuth(launcher.Ec2Pool.AccessKey, launcher.Ec2Pool.SecretKey, "", time.Time{})
	if err != nil {
		panic("Aws credentials were null")
	}

	return launcher.ec2Broker.Ec2Cache(auth)
}

func (launcher *Ec2VirtualMachineLauncher) LaunchVirtualMachine() (vm.VirtualMachine, error) {
	username := launcher.Ec2Pool.Username
	activeImage, err := launcher.getActiveImage()
	if err != nil {
		return nil, err
	}

	securityGroups, err := launcher.getSecurityGroups()
	if err != nil {
		return nil, err
	}

	runOptions := ec2.RunInstancesOptions{
		ImageId:        activeImage.Id,
		InstanceType:   launcher.Ec2Pool.InstanceType,
		SecurityGroups: securityGroups,
		UserData:       launcher.getUserData(username),
	}
	runResponse, err := launcher.ec2Cache().EC2.RunInstances(&runOptions)
	if err != nil {
		return nil, err
	}

	if len(runResponse.Instances) == 0 {
		return nil, errors.New("No instances launched")
	}
	if len(runResponse.Instances) > 1 {
		fmt.Printf("Launched too many instances. Wat?")
		extraInstanceIds := make([]string, len(runResponse.Instances)-1)
		for i := 1; i < len(runResponse.Instances); i++ {
			extraInstanceIds[i-1] = runResponse.Instances[i].InstanceId
		}
		_, err := launcher.ec2Cache().EC2.TerminateInstances(extraInstanceIds)
		if err != nil {
			launcher.ec2Cache().EC2.TerminateInstances([]string{runResponse.Instances[0].InstanceId})
			return nil, err
		}
	}
	instance := &runResponse.Instances[0]
	nameTag := ec2.Tag{"Name", fmt.Sprintf("koality-worker (%s)", launcher.ec2Broker.InstanceInfo().Name)}
	_, err = launcher.ec2Cache().EC2.CreateTags([]string{instance.InstanceId}, []ec2.Tag{nameTag})
	err = launcher.waitForIpAddress(instance, 2*time.Minute)
	if err != nil {
		return nil, err
	}
	ec2Vm, err := launcher.waitForSsh(instance, username, 3*time.Minute)
	if err != nil {
		return nil, err
	}
	return ec2Vm, nil
}

func (launcher *Ec2VirtualMachineLauncher) waitForIpAddress(instance *ec2.Instance, timeout time.Duration) error {
	for {
		select {
		case <-time.After(timeout):
			_, err := launcher.ec2Cache().EC2.TerminateInstances([]string{instance.InstanceId})
			if err != nil {
				return err
			}
			return fmt.Errorf("Instance failed to receive an IP address after %s", timeout.String())
		default:
			if instance.PrivateIPAddress != "" {
				return nil
			} else {
				time.Sleep(5 * time.Second)
				for _, reservation := range launcher.ec2Cache().Reservations() {
					for _, inst := range reservation.Instances {
						if inst.InstanceId == instance.InstanceId {
							*instance = inst
							break
						}
					}
				}
			}
		}
	}
}

func (launcher *Ec2VirtualMachineLauncher) waitForSsh(instance *ec2.Instance, username string, timeout time.Duration) (*Ec2VirtualMachine, error) {
	for {
		select {
		case <-time.After(timeout):
			_, err := launcher.ec2Cache().EC2.TerminateInstances([]string{instance.InstanceId})
			if err != nil {
				return new(Ec2VirtualMachine), err
			}
			return nil, fmt.Errorf("Failed to ssh into the instance after %s", timeout.String())
		default:
			ec2Vm, err := New(instance, launcher.ec2Cache(), username)
			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}
			sshAttempt, err := ec2Vm.MakeExecutable(shell.Command("true"), nil, nil, nil, nil)
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
func (launcher *Ec2VirtualMachineLauncher) getActiveImage() (ec2.Image, error) {
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

func (launcher *Ec2VirtualMachineLauncher) getSnapshotVersion(snapshot ec2.Image) (int, error) {
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

func (launcher *Ec2VirtualMachineLauncher) getBaseImage() (ec2.Image, error) {
	if launcher.Ec2Pool.BaseAmiId != "" {
		imagesResponse, err := launcher.ec2Cache().EC2.Images([]string{launcher.Ec2Pool.BaseAmiId}, nil)
		if err != nil {
			return ec2.Image{}, err
		}

		if len(imagesResponse.Images) > 0 {
			return imagesResponse.Images[0], nil
		}
	}
	imageFilter := ec2.NewFilter()
	imageFilter.Add("owner-id", "600991114254") // must be changed if our ec2 info changes
	imageFilter.Add("name", "koality_verification_precise_0.4")
	imageFilter.Add("state", "available")
	imagesResponse, err := launcher.ec2Cache().EC2.Images(nil, imageFilter)
	if err != nil {
		return ec2.Image{}, err
	}
	return imagesResponse.Images[0], nil
}

func (launcher *Ec2VirtualMachineLauncher) getSnapshotsForImage(baseImage ec2.Image) ([]ec2.Image, error) {
	imageFilter := ec2.NewFilter()
	imageFilter.Add("name", fmt.Sprintf("koality-snapshot-(%s/%s)-v*", launcher.Ec2Pool.Name, baseImage.Name))
	imageFilter.Add("state", "available")
	imagesResponse, err := launcher.ec2Cache().EC2.Images(nil, imageFilter)
	if err != nil {
		return nil, err
	}
	return imagesResponse.Images, nil
}

func (launcher *Ec2VirtualMachineLauncher) getSecurityGroups() ([]ec2.SecurityGroup, error) {
	var securityGroup ec2.SecurityGroup
	if launcher.Ec2Pool.SecurityGroupId == "koality_verification" {
		securityGroup = ec2.SecurityGroup{
			Name: "koality_verification",
		}
	} else {
		securityGroup = ec2.SecurityGroup{
			Id: launcher.Ec2Pool.SecurityGroupId,
		}
	}

	securityGroupsResp, err := launcher.ec2Cache().EC2.SecurityGroups([]ec2.SecurityGroup{securityGroup}, nil)
	if err != nil {
		return nil, err
	}

	if len(securityGroupsResp.Groups) == 0 {
		securityGroup = ec2.SecurityGroup{
			Name: "koality_verification",
		}
		_, err := launcher.ec2Cache().EC2.CreateSecurityGroup("koality_verification", "Auto-generated security group which allows the Koality master to ssh into its launched testing instances.")
		if err != nil {
			return nil, err
		}
	} else {
		groupInfo := securityGroupsResp.Groups[0]
		for _, ipPerm := range groupInfo.IPPerms {
			if ipPerm.FromPort <= 22 && 22 <= ipPerm.ToPort {
				for _, sourceGroup := range ipPerm.SourceGroups {
					for _, ownGroup := range launcher.ec2Broker.InstanceInfo().SecurityGroups {
						if ownGroup.Id == sourceGroup.Id {
							return []ec2.SecurityGroup{securityGroup}, nil
						}
					}
				}
				for _, sourceIPRange := range ipPerm.SourceIPs {
					_, ipNet, err := net.ParseCIDR(sourceIPRange)
					if err != nil {
						return nil, err
					}

					if ipNet.Contains(launcher.ec2Broker.InstanceInfo().PrivateIp) {
						return []ec2.SecurityGroup{securityGroup}, nil
					}
				}
			}
		}
	}

	ownIp := launcher.ec2Broker.InstanceInfo().PrivateIp
	ipNet := net.IPNet{ownIp, ownIp.DefaultMask()}
	sshPerm := ec2.IPPerm{
		Protocol:  "tcp",
		FromPort:  22,
		ToPort:    22,
		SourceIPs: []string{ipNet.String()},
	}
	_, err = launcher.ec2Cache().EC2.AuthorizeSecurityGroup(securityGroup, []ec2.IPPerm{sshPerm})
	if err != nil {
		return nil, err
	}

	return []ec2.SecurityGroup{securityGroup}, nil
}

func (launcher *Ec2VirtualMachineLauncher) getUserData(username string) []byte {
	buffer := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(buffer)
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\nMIME-Version: 1.0\n\n", multipartWriter.Boundary()))

	err := writeCloudInitMimePart(multipartWriter, launcher.getDefaultUserData(username), "koality-default-data")
	if err != nil {
		panic(err)
	}

	customUserData := launcher.Ec2Pool.UserData
	if customUserData != "" {
		err := writeCloudInitMimePart(multipartWriter, customUserData, "koality-custom-data")
		if err != nil {
			panic(err)
		}
	}

	multipartWriter.Close()

	return buffer.Bytes()
}

func (launcher *Ec2VirtualMachineLauncher) getDefaultUserData(username string) string {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	keyBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/.ssh/id_rsa.pub", currentUser.HomeDir))
	if err != nil {
		panic(err)
	}
	publicKey := string(keyBytes)

	configureUserCommand := shell.Chain(
		shell.Commandf("useradd --create-home -s /bin/bash %s", username),
		shell.Commandf("mkdir ~%s/.ssh", username),
		shell.Append(shell.Commandf("echo %s", shell.Quote(publicKey)), shell.Commandf("~%s/.ssh/authorized_keys", username), false),
		shell.Commandf("chown -R %s:%s ~%s/.ssh", username, username, username),
		shell.Or(
			shell.Command("grep '#includedir /etc/sudoers.d' /etc/sudoers"),
			shell.Append(shell.Command("echo #includedir /etc/sudoers.d"), shell.Command("/etc/sudoers"), false),
		),
		shell.Command("mkdir /etc/sudoers.d"),
		shell.Redirect(shell.Commandf("echo 'Defaults !requiretty\n%s ALL=(ALL) NOPASSWD: ALL'", username), shell.Commandf("/etc/sudoers.d/koality-%s", username), false),
		shell.Commandf("chmod 0440 /etc/sudoers.d/koality-%s", username),
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
