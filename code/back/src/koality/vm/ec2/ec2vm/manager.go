package ec2vm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	"koality/resources"
	"koality/shell"
	"koality/vm"
	"koality/vm/cloudinit"
	"koality/vm/ec2/ec2broker"
	"net"
	"path"
	"strconv"
	"strings"
	"time"
)

type Ec2VirtualMachineManager struct {
	ec2Broker           *ec2broker.Ec2Broker
	Ec2Pool             *resources.Ec2Pool
	ec2Cache            *ec2broker.Ec2Cache
	resourcesConnection *resources.Connection
}

func NewManager(ec2Broker *ec2broker.Ec2Broker, ec2Pool *resources.Ec2Pool, resourcesConnection *resources.Connection) (*Ec2VirtualMachineManager, error) {
	auth, err := aws.GetAuth(ec2Pool.AccessKey, ec2Pool.SecretKey, "", time.Time{})
	if err != nil {
		return nil, err
	}

	ec2Cache, err := ec2Broker.Ec2Cache(auth)
	if err != nil {
		return nil, err
	}

	return &Ec2VirtualMachineManager{ec2Broker, ec2Pool, ec2Cache, resourcesConnection}, nil
}

func (manager *Ec2VirtualMachineManager) NewVirtualMachine() (vm.VirtualMachine, error) {
	username := manager.Ec2Pool.Username
	activeImage, err := manager.getActiveImage()
	if err != nil {
		return nil, err
	}

	securityGroups, err := manager.getSecurityGroups()
	if err != nil {
		return nil, err
	}

	keyPair, err := manager.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		return nil, err
	}

	userData, err := manager.getUserData(username, keyPair)
	if err != nil {
		return nil, err
	}

	runOptions := ec2.RunInstancesOptions{
		ImageId:             activeImage.Id,
		InstanceType:        manager.Ec2Pool.InstanceType,
		SecurityGroups:      securityGroups,
		SubnetId:            manager.Ec2Pool.VpcSubnetId,
		UserData:            userData,
		BlockDeviceMappings: manager.getBlockDeviceMappings(activeImage),
	}
	runResponse, err := manager.ec2Cache.EC2.RunInstances(&runOptions)
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
		_, err := manager.ec2Cache.EC2.TerminateInstances(extraInstanceIds)
		if err != nil {
			manager.ec2Cache.EC2.TerminateInstances([]string{runResponse.Instances[0].InstanceId})
			return nil, err
		}
	}
	instance := runResponse.Instances[0]
	nameTag := ec2.Tag{"Name", fmt.Sprintf("koality-worker (%s)", manager.ec2Broker.InstanceInfo().Name)}
	_, err = manager.ec2Cache.EC2.CreateTags([]string{instance.InstanceId}, []ec2.Tag{nameTag})
	if err != nil {
		manager.ec2Cache.EC2.TerminateInstances([]string{instance.InstanceId})
		return nil, err
	}

	err = manager.waitForIpAddress(&instance, 2*time.Minute)
	if err != nil {
		return nil, err
	}

	ec2Vm, err := manager.waitForSsh(instance, username, keyPair, 5*time.Minute)
	if err != nil {
		manager.ec2Cache.EC2.TerminateInstances([]string{instance.InstanceId})
		return nil, err
	}
	return ec2Vm, nil
}

func (manager *Ec2VirtualMachineManager) GetVirtualMachine(instanceId string) (vm.VirtualMachine, error) {
	keyPair, err := manager.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		return nil, err
	}

	instancesResp, err := manager.ec2Cache.EC2.Instances([]string{instanceId}, nil)
	if err != nil {
		return nil, err
	} else if len(instancesResp.Reservations) == 0 || len(instancesResp.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("No instances found for id %s", instanceId)
	}

	instance := instancesResp.Reservations[0].Instances[0]
	return New(instance, manager.ec2Cache, manager.Ec2Pool.Username, keyPair.PrivateKey)
}

func (manager *Ec2VirtualMachineManager) waitForIpAddress(instance *ec2.Instance, timeout time.Duration) error {
	for {
		select {
		case <-time.After(timeout):
			_, err := manager.ec2Cache.EC2.TerminateInstances([]string{instance.InstanceId})
			if err != nil {
				return err
			}
			return fmt.Errorf("Instance failed to receive an IP address after %s", timeout.String())
		default:
			if instance.PrivateIPAddress != "" {
				return nil
			} else {
				time.Sleep(5 * time.Second)
				for _, reservation := range manager.ec2Cache.Reservations() {
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

func (manager *Ec2VirtualMachineManager) waitForSsh(instance ec2.Instance, username string, keyPair *resources.RepositoryKeyPair, timeout time.Duration) (*Ec2VirtualMachine, error) {
	for {
		select {
		case <-time.After(timeout):
			_, err := manager.ec2Cache.EC2.TerminateInstances([]string{instance.InstanceId})
			if err != nil {
				return new(Ec2VirtualMachine), err
			}
			return nil, fmt.Errorf("Failed to ssh into the instance after %s", timeout.String())
		default:
			ec2Vm, err := New(instance, manager.ec2Cache, username, keyPair.PrivateKey)
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

func (manager *Ec2VirtualMachineManager) getActiveImage() (ec2.Image, error) {
	baseImage, err := manager.getBaseImage()
	if err != nil {
		return ec2.Image{}, err
	}
	snapshots, err := manager.getSnapshotsForImage(baseImage)
	if err != nil || len(snapshots) == 0 {
		return baseImage, nil
	}
	newestSnapshot := snapshots[0]
	newestSnapshotVersion, _ := manager.getSnapshotVersion(newestSnapshot)
	for _, snapshot := range snapshots {
		snapshotVersion, err := manager.getSnapshotVersion(snapshot)
		if err != nil && snapshotVersion > newestSnapshotVersion {
			newestSnapshotVersion = snapshotVersion
			newestSnapshot = snapshot
		}
	}
	return newestSnapshot, nil
}

func (manager *Ec2VirtualMachineManager) getSnapshotVersion(snapshot ec2.Image) (int, error) {
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

func (manager *Ec2VirtualMachineManager) getBaseImage() (ec2.Image, error) {
	if manager.Ec2Pool.BaseAmiId != "" {
		imagesResponse, err := manager.ec2Cache.EC2.Images([]string{manager.Ec2Pool.BaseAmiId}, nil)
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
	imagesResponse, err := manager.ec2Cache.EC2.Images(nil, imageFilter)
	if err != nil {
		return ec2.Image{}, err
	}
	return imagesResponse.Images[0], nil
}

func (manager *Ec2VirtualMachineManager) getSnapshotsForImage(baseImage ec2.Image) ([]ec2.Image, error) {
	imageFilter := ec2.NewFilter()
	imageFilter.Add("name", fmt.Sprintf("koality-snapshot-(%s/%s)-v*", manager.Ec2Pool.Name, baseImage.Name))
	imageFilter.Add("state", "available")
	imagesResponse, err := manager.ec2Cache.EC2.Images(nil, imageFilter)
	if err != nil {
		return nil, err
	}
	return imagesResponse.Images, nil
}

func (manager *Ec2VirtualMachineManager) getSecurityGroups() ([]ec2.SecurityGroup, error) {
	var securityGroup ec2.SecurityGroup
	var securityGroups []ec2.SecurityGroup
	defaultSecurityGroup := ec2.SecurityGroup{
		Name: "koality_verification",
	}
	if manager.Ec2Pool.SecurityGroupId == "" {
		securityGroup = defaultSecurityGroup
		securityGroups = []ec2.SecurityGroup{defaultSecurityGroup}
	} else {
		securityGroup = ec2.SecurityGroup{
			Id: manager.Ec2Pool.SecurityGroupId,
		}
		securityGroups = []ec2.SecurityGroup{
			securityGroup,
			defaultSecurityGroup,
		}
	}

	securityGroupsResp, err := manager.ec2Cache.EC2.SecurityGroups(securityGroups, nil)
	if err != nil {
		return nil, err
	}

	if len(securityGroupsResp.Groups) == 0 {
		securityGroup = ec2.SecurityGroup{
			Name: "koality_verification",
		}
		_, err := manager.ec2Cache.EC2.CreateSecurityGroup("koality_verification", "Auto-generated security group which allows the Koality master to ssh into its launched testing instances.")
		if err != nil {
			return nil, err
		}
	} else {
		groupInfo := securityGroupsResp.Groups[0]
		if len(securityGroupsResp.Groups) == 2 {
			for _, group := range securityGroupsResp.Groups {
				if group.Id == securityGroup.Id {
					groupInfo = group
					break
				}
			}
		}
		for _, ipPerm := range groupInfo.IPPerms {
			if ipPerm.FromPort <= 22 && 22 <= ipPerm.ToPort {
				for _, sourceGroup := range ipPerm.SourceGroups {
					for _, ownGroup := range manager.ec2Broker.InstanceInfo().SecurityGroups {
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

					if ipNet.Contains(manager.ec2Broker.InstanceInfo().PrivateIp) {
						return []ec2.SecurityGroup{securityGroup}, nil
					}
				}
			}
		}
	}

	ownIp := manager.ec2Broker.InstanceInfo().PrivateIp
	ipNet := net.IPNet{ownIp, ownIp.DefaultMask()}
	sshPerm := ec2.IPPerm{
		Protocol:  "tcp",
		FromPort:  22,
		ToPort:    22,
		SourceIPs: []string{ipNet.String()},
	}
	_, err = manager.ec2Cache.EC2.AuthorizeSecurityGroup(securityGroup, []ec2.IPPerm{sshPerm})
	if err != nil {
		return nil, err
	}

	return []ec2.SecurityGroup{securityGroup}, nil
}

func (manager *Ec2VirtualMachineManager) getBlockDeviceMappings(image ec2.Image) []ec2.BlockDeviceMapping {
	rootDriveSize := int64(manager.Ec2Pool.RootDriveSize)

	blockDeviceMappings := make([]ec2.BlockDeviceMapping, 0, len(image.BlockDevices))
	for _, blockDeviceMapping := range image.BlockDevices {
		if blockDeviceMapping.DeviceName == image.RootDeviceName && rootDriveSize > blockDeviceMapping.VolumeSize {
			rootDriveMapping := ec2.BlockDeviceMapping{
				DeviceName:          blockDeviceMapping.DeviceName,
				VirtualName:         blockDeviceMapping.VirtualName,
				SnapshotId:          blockDeviceMapping.SnapshotId,
				VolumeType:          blockDeviceMapping.VolumeType,
				VolumeSize:          rootDriveSize,
				DeleteOnTermination: blockDeviceMapping.DeleteOnTermination,
				IOPS:                blockDeviceMapping.IOPS,
			}
			blockDeviceMappings = append(blockDeviceMappings, rootDriveMapping)
		} else {
			blockDeviceMappings = append(blockDeviceMappings, blockDeviceMapping)
		}
	}
	return blockDeviceMappings
}

func (manager *Ec2VirtualMachineManager) getUserData(username string, keyPair *resources.RepositoryKeyPair) ([]byte, error) {
	buffer := new(bytes.Buffer)
	mimeMultipartWriter := cloudinit.NewMimeMultipartWriter(buffer)

	defaultUserData := manager.getDefaultUserData(username, keyPair)

	if err := mimeMultipartWriter.WriteMimePart("koality-default-data", defaultUserData); err != nil {
		return nil, err
	}

	customUserData := strings.TrimSpace(manager.Ec2Pool.UserData)
	if customUserData != "" {
		if err := mimeMultipartWriter.WriteMimePart("koality-custom-data", customUserData); err != nil {
			return nil, err
		}
	}

	if err := mimeMultipartWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (manager *Ec2VirtualMachineManager) getDefaultUserData(username string, keyPair *resources.RepositoryKeyPair) string {
	configureUserCommand := func(username string) shell.Command {
		homeDir := "~" + username
		sshDir := path.Join(homeDir, ".ssh")
		privateKeyPath := path.Join(sshDir, "id_rsa")
		publicKeyPath := path.Join(sshDir, "id_rsa.pub")
		authorizedKeysPath := path.Join(sshDir, "authorized_keys")
		sshConfigPath := path.Join(sshDir, "config")
		sshConfigContents := "Host *\n  StrictHostKeyChecking no"

		return shell.Chain(
			shell.Commandf("mkdir -p %s", sshDir),
			shell.Redirect(shell.Commandf("echo %s", shell.Quote(keyPair.PrivateKey)), shell.Command(privateKeyPath), false),
			shell.Redirect(shell.Commandf("echo %s", shell.Quote(keyPair.PublicKey)), shell.Command(publicKeyPath), false),
			shell.Append(shell.Commandf("echo %s", shell.Quote(keyPair.PublicKey)), shell.Command(authorizedKeysPath), false),
			shell.Or(
				shell.Commandf("grep %s %s", shell.Quote(sshConfigContents), shell.Command(sshConfigPath)),
				shell.Append(shell.Commandf("echo %s", shell.Quote(sshConfigContents)), shell.Command(sshConfigPath), false),
			),
			shell.Commandf("chmod -R 0400 %s", sshDir),
			shell.Commandf("chmod 0700 %s", sshDir),
			shell.Commandf("chown -R %s:%s %s", username, username, sshDir),
		)
	}

	sudoersFilePath := path.Join("/", "etc", "sudoers")
	sudoersDirPath := path.Join("/", "etc", "sudoers.d")
	sudoersUserFilePath := path.Join(sudoersDirPath, fmt.Sprintf("koality-%s", username))

	addUserCommand := shell.Commandf("useradd --create-home -s /bin/bash %s", username)

	configureSudoersCommand := shell.Chain(
		shell.Or(
			shell.Commandf("grep %s %s", shell.Quote(fmt.Sprintf("#includedir %s", sudoersDirPath)), shell.Command(sudoersFilePath)),
			shell.Append(shell.Commandf("echo %s", shell.Quote(fmt.Sprintf("#includedir %s", sudoersDirPath))), shell.Command(sudoersFilePath), false),
		),
		shell.Commandf("mkdir -p %s", sudoersDirPath),
		shell.Redirect(shell.Commandf("echo 'Defaults !requiretty\n%s ALL=(ALL) NOPASSWD: ALL'", username), shell.Command(sudoersUserFilePath), false),
		shell.Commandf("chmod 0440 %s", sudoersUserFilePath),
	)

	configureInstanceCommand := shell.Chain(
		addUserCommand,
		configureUserCommand(username),
		configureSudoersCommand,
	)

	if username != "root" {
		configureInstanceCommand = shell.Chain(
			configureInstanceCommand,
			configureUserCommand("root"),
		)
	}

	return fmt.Sprintf("#!/bin/sh\n%s", configureInstanceCommand)
}
