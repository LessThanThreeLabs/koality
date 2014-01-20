package database

import (
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidEc2Pool(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	name := "ec2-pool"
	accessKey := "aaaabbbbccccddddeeee"
	secretKey := "0000111122223333444455556666777788889999"
	username := "koality"
	baseAmiId := "ami-12345678"
	securityGroupId := "sg-12345678"
	vpcSubnetId := "subnet-12345678"
	instanceType := "m1.medium"
	numReadyInstances := uint64(2)
	numMaxInstances := uint64(10)
	rootDriveSize := uint64(100)
	userData := "echo hello"

	_, err = connection.Pools.Create.CreateEc2Pool("", accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid name")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, "zzz", secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid access key")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, "123", username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid secret key")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, "", baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid username")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, "bad-ami-id", securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid base ami id")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, "asdf", vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid security group id")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, "asdf",
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid vpc subnet id")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		"not-an-instance-type", numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid instance type")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, 100, 99, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing numReadyInstances > numMaxInstances")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, 0, 0, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid number of max instances")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, 0, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid root drive size")
	}
}

func TestCreateEc2Pool(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	poolCreatedEventReceived := make(chan bool, 1)
	var poolCreatedEventEc2Pool *resources.Ec2Pool
	poolEc2CreatedHandler := func(ec2Pool *resources.Ec2Pool) {
		poolCreatedEventEc2Pool = ec2Pool
		poolCreatedEventReceived <- true
	}
	_, err = connection.Pools.Subscription.SubscribeToEc2CreatedEvents(poolEc2CreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	poolDeletedEventReceived := make(chan bool, 1)
	poolDeletedEventId := uint64(0)
	poolEc2DeletedHandler := func(ec2PoolId uint64) {
		poolDeletedEventId = ec2PoolId
		poolDeletedEventReceived <- true
	}
	_, err = connection.Pools.Subscription.SubscribeToEc2DeletedEvents(poolEc2DeletedHandler)
	if err != nil {
		test.Fatal(err)
	}

	name := "ec2-pool"
	accessKey := "aaaabbbbccccddddeeee"
	secretKey := "0000111122223333444455556666777788889999"
	username := "koality"
	baseAmiId := "ami-12345678"
	securityGroupId := "sg-12345678"
	vpcSubnetId := "subnet-12345678"
	instanceType := "m1.medium"
	numReadyInstances := uint64(2)
	numMaxInstances := uint64(10)
	rootDriveSize := uint64(100)
	userData := "echo hello"

	pool1, err := connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		test.Fatal(err)
	}

	if pool1.Name != name {
		test.Fatal("pool.Name mismatch")
	} else if pool1.AccessKey != accessKey {
		test.Fatal("pool.AccessKey mismatch")
	} else if pool1.SecretKey != secretKey {
		test.Fatal("pool.SecretKey mismatch")
	} else if pool1.Username != username {
		test.Fatal("pool.Username mismatch")
	} else if pool1.BaseAmiId != baseAmiId {
		test.Fatal("pool.BaseAmiId mismatch")
	} else if pool1.SecurityGroupId != securityGroupId {
		test.Fatal("pool.SecurityGroupId mismatch")
	} else if pool1.VpcSubnetId != vpcSubnetId {
		test.Fatal("pool.VpcSubnetId mismatch")
	} else if pool1.InstanceType != instanceType {
		test.Fatal("pool.InstanceType mismatch")
	} else if pool1.NumReadyInstances != numReadyInstances {
		test.Fatal("pool.NumReadyInstances mismatch")
	} else if pool1.NumMaxInstances != numMaxInstances {
		test.Fatal("pool.NumMaxInstances mismatch")
	} else if pool1.RootDriveSize != rootDriveSize {
		test.Fatal("pool.RootDriveSize mismatch")
	} else if pool1.UserData != userData {
		test.Fatal("pool.UserData mismatch")
	}

	select {
	case <-poolCreatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear ec2 pool creation event")
	}

	if poolCreatedEventEc2Pool.Id != pool1.Id {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.Name != pool1.Name {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.AccessKey != pool1.AccessKey {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.SecretKey != pool1.SecretKey {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.Username != pool1.Username {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.BaseAmiId != pool1.BaseAmiId {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.SecurityGroupId != pool1.SecurityGroupId {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.VpcSubnetId != pool1.VpcSubnetId {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.InstanceType != pool1.InstanceType {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.NumReadyInstances != pool1.NumReadyInstances {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.NumMaxInstances != pool1.NumMaxInstances {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.RootDriveSize != pool1.RootDriveSize {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	} else if poolCreatedEventEc2Pool.UserData != pool1.UserData {
		test.Fatal("Bad pool.Id in ec2 pool creation event")
	}

	pool1Again, err := connection.Pools.Read.GetEc2Pool(pool1.Id)
	if err != nil {
		test.Fatal(err)
	}

	if pool1.Name != pool1Again.Name {
		test.Fatal("pool.Name mismatch")
	} else if pool1.AccessKey != pool1Again.AccessKey {
		test.Fatal("pool.AccessKey mismatch")
	} else if pool1.SecretKey != pool1Again.SecretKey {
		test.Fatal("pool.SecretKey mismatch")
	} else if pool1.Username != pool1Again.Username {
		test.Fatal("pool.Username mismatch")
	} else if pool1.BaseAmiId != pool1Again.BaseAmiId {
		test.Fatal("pool.BaseAmiId mismatch")
	} else if pool1.SecurityGroupId != pool1Again.SecurityGroupId {
		test.Fatal("pool.SecurityGroupId mismatch")
	} else if pool1.VpcSubnetId != pool1Again.VpcSubnetId {
		test.Fatal("pool.VpcSubnetId mismatch")
	} else if pool1.InstanceType != pool1Again.InstanceType {
		test.Fatal("pool.InstanceType mismatch")
	} else if pool1.NumReadyInstances != pool1Again.NumReadyInstances {
		test.Fatal("pool.NumReadyInstances mismatch")
	} else if pool1.NumMaxInstances != pool1Again.NumMaxInstances {
		test.Fatal("pool.NumMaxInstances mismatch")
	} else if pool1.RootDriveSize != pool1Again.RootDriveSize {
		test.Fatal("pool.RootDriveSize mismatch")
	} else if pool1.UserData != pool1Again.UserData {
		test.Fatal("pool.UserData mismatch")
	}

	pools, err := connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	if len(pools) != 1 {
		test.Fatal("Expected there to only be one pool")
	}

	pool2, err := connection.Pools.Create.CreateEc2Pool(name+"2", accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		test.Fatal(err)
	}

	pool2Again, err := connection.Pools.Read.GetEc2Pool(pool2.Id)
	if err != nil {
		test.Fatal(err)
	}

	if pool2.Name != pool2Again.Name {
		test.Fatal("pool.Name mismatch")
	} else if pool2.AccessKey != pool2Again.AccessKey {
		test.Fatal("pool.AccessKey mismatch")
	} else if pool2.SecretKey != pool2Again.SecretKey {
		test.Fatal("pool.SecretKey mismatch")
	} else if pool2.Username != pool2Again.Username {
		test.Fatal("pool.Username mismatch")
	} else if pool2.BaseAmiId != pool2Again.BaseAmiId {
		test.Fatal("pool.BaseAmiId mismatch")
	} else if pool2.SecurityGroupId != pool2Again.SecurityGroupId {
		test.Fatal("pool.SecurityGroupId mismatch")
	} else if pool2.VpcSubnetId != pool2Again.VpcSubnetId {
		test.Fatal("pool.VpcSubnetId mismatch")
	} else if pool2.InstanceType != pool2Again.InstanceType {
		test.Fatal("pool.InstanceType mismatch")
	} else if pool2.NumReadyInstances != pool2Again.NumReadyInstances {
		test.Fatal("pool.NumReadyInstances mismatch")
	} else if pool2.NumMaxInstances != pool2Again.NumMaxInstances {
		test.Fatal("pool.NumMaxInstances mismatch")
	} else if pool2.RootDriveSize != pool2Again.RootDriveSize {
		test.Fatal("pool.RootDriveSize mismatch")
	} else if pool2.UserData != pool2Again.UserData {
		test.Fatal("pool.UserData mismatch")
	}

	pools, err = connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	if len(pools) != 2 {
		test.Fatal("Expected there to be two pools")
	}

	err = connection.Pools.Delete.DeleteEc2Pool(pool1.Id)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-poolDeletedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear ec2 pool deletion event")
	}

	if poolDeletedEventId != pool1.Id {
		test.Fatal("Bad poolId in ec2 pool deletion event")
	}

	err = connection.Pools.Delete.DeleteEc2Pool(pool1.Id)
	if _, ok := err.(resources.NoSuchPoolError); !ok {
		test.Fatal("Expected NoSuchPoolError when trying to delete same pool twice")
	}

	pools, err = connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	if len(pools) != 1 {
		test.Fatal("Expected there to be only one pool")
	}
}

func TestUsersEc2Settings(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	poolEventReceived := make(chan bool, 1)
	poolEventId := uint64(0)
	poolEventAccessKey := ""
	poolEventSecretKey := ""
	poolEventUsername := ""
	poolEventBaseAmiId := ""
	poolEventSecurityGroupId := ""
	poolEventVpcSubnetId := ""
	poolEventInstanceType := ""
	poolEventNumReadyInstances := uint64(0)
	poolEventNumMaxInstances := uint64(0)
	poolEventRootDriveSize := uint64(0)
	poolEventUserData := ""
	ec2PoolSettingsUpdatedHandler := func(ec2PoolId uint64, accessKey, secretKey,
		username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
		numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) {
		poolEventId = ec2PoolId
		poolEventAccessKey = accessKey
		poolEventSecretKey = secretKey
		poolEventUsername = username
		poolEventBaseAmiId = baseAmiId
		poolEventSecurityGroupId = securityGroupId
		poolEventVpcSubnetId = vpcSubnetId
		poolEventInstanceType = instanceType
		poolEventNumReadyInstances = numReadyInstances
		poolEventNumMaxInstances = numMaxInstances
		poolEventRootDriveSize = rootDriveSize
		poolEventUserData = userData
		poolEventReceived <- true
	}
	_, err = connection.Pools.Subscription.SubscribeToEc2SettingsUpdatedEvents(ec2PoolSettingsUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	name := "ec2-pool"
	accessKey := "aaaabbbbccccddddeeee"
	accessKey2 := "eeeeddddccccbbbbaaaa"
	secretKey := "0000111122223333444455556666777788889999"
	secretKey2 := "9999888877776666555544443333222211110000"
	username := "koality"
	username2 := "koality2"
	baseAmiId := "ami-12345678"
	baseAmiId2 := "ami-87654321"
	securityGroupId := "sg-12345678"
	securityGroupId2 := "sg-87654321"
	vpcSubnetId := "subnet-12345678"
	vpcSubnetId2 := "subnet-87654321"
	instanceType := "m1.medium"
	instanceType2 := "m1.large"
	numReadyInstances := uint64(2)
	numReadyInstances2 := uint64(4)
	numMaxInstances := uint64(10)
	numMaxInstances2 := uint64(100)
	rootDriveSize := uint64(100)
	rootDriveSize2 := uint64(150)
	userData := "echo hello"
	userData2 := "echo hello 2"

	pool, err := connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Pools.Update.SetEc2Settings(pool.Id, accessKey2, secretKey2, username2, baseAmiId2, securityGroupId2, vpcSubnetId2,
		instanceType2, numReadyInstances2, numMaxInstances2, rootDriveSize2, userData2)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-poolEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear ec2 pool settings updated updated event")
	}

	if poolEventId != pool.Id {
		test.Fatal("Bad poolId in ec2 pool settings updated event")
	} else if poolEventAccessKey != accessKey2 {
		test.Fatal("Bad access key in ec2 pool settings updated event")
	} else if poolEventSecretKey != secretKey2 {
		test.Fatal("Bad secret key in ec2 pool settings updated event")
	} else if poolEventUsername != username2 {
		test.Fatal("Bad username in ec2 pool settings updated event")
	} else if poolEventBaseAmiId != baseAmiId2 {
		test.Fatal("Bad base ami id in ec2 pool settings updated event")
	} else if poolEventSecurityGroupId != securityGroupId2 {
		test.Fatal("Bad security group id in ec2 pool settings updated event")
	} else if poolEventVpcSubnetId != vpcSubnetId2 {
		test.Fatal("Bad vpc subnet id in ec2 pool settings updated event")
	} else if poolEventInstanceType != instanceType2 {
		test.Fatal("Bad instance type in ec2 pool settings updated event")
	} else if poolEventNumReadyInstances != numReadyInstances2 {
		test.Fatal("Bad num ready instances in ec2 pool settings updated event")
	} else if poolEventNumMaxInstances != numMaxInstances2 {
		test.Fatal("Bad num max instances in ec2 pool settings updated event")
	} else if poolEventRootDriveSize != rootDriveSize2 {
		test.Fatal("Bad root drive size in ec2 pool settings updated event")
	} else if poolEventUserData != userData2 {
		test.Fatal("Bad user data in ec2 pool settings updated event")
	}

	pool2, err := connection.Pools.Read.GetEc2Pool(pool.Id)
	if err != nil {
		test.Fatal(err)
	}

	if pool2.AccessKey != accessKey2 {
		test.Fatal("pool.AccessKey mismatch")
	} else if pool2.SecretKey != secretKey2 {
		test.Fatal("pool.SecretKey mismatch")
	} else if pool2.Username != username2 {
		test.Fatal("pool.Username mismatch")
	} else if pool2.BaseAmiId != baseAmiId2 {
		test.Fatal("pool.BaseAmiId mismatch")
	} else if pool2.SecurityGroupId != securityGroupId2 {
		test.Fatal("pool.SecurityGroupId mismatch")
	} else if pool2.VpcSubnetId != vpcSubnetId2 {
		test.Fatal("pool.VpcSubnetId mismatch")
	} else if pool2.InstanceType != instanceType2 {
		test.Fatal("pool.InstanceType mismatch")
	} else if pool2.NumReadyInstances != numReadyInstances2 {
		test.Fatal("pool.NumReadyInstances mismatch")
	} else if pool2.NumMaxInstances != numMaxInstances2 {
		test.Fatal("pool.NumMaxInstances mismatch")
	} else if pool2.RootDriveSize != rootDriveSize2 {
		test.Fatal("pool.RootDriveSize mismatch")
	} else if pool2.UserData != userData2 {
		test.Fatal("pool.UserData mismatch")
	}

	err = connection.Pools.Update.SetEc2Settings(0, accessKey2, secretKey2, username2, baseAmiId2, securityGroupId2, vpcSubnetId2,
		instanceType2, numReadyInstances2, numMaxInstances2, rootDriveSize2, userData2)
	if _, ok := err.(resources.NoSuchPoolError); !ok {
		test.Fatal("Expected NoSuchPoolError when trying to update nonexistent ec2 pool")
	}
}
