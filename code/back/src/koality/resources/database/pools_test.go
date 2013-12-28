package database

import (
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidEc2Pool(test *testing.T) {
	PopulateDatabase()

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

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, "", vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err == nil {
		test.Fatal("Expected error after providing invalid security group id")
	}

	_, err = connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, "",
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
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	poolCreatedEventReceived := make(chan bool, 1)
	poolCreatedEventId := uint64(0)
	poolEc2CreatedHandler := func(ec2PoolId uint64) {
		poolCreatedEventId = ec2PoolId
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

	pool1Id, err := connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-poolCreatedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear ec2 pool creation event")
	}

	if poolCreatedEventId != pool1Id {
		test.Fatal("Bad poolId in ec2 pool creation event")
	}

	pool1, err := connection.Pools.Read.GetEc2Pool(pool1Id)
	if err != nil {
		test.Fatal(err)
	}

	if pool1.Id != pool1Id {
		test.Fatal("pool.Id mismatch")
	}

	pools, err := connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	if len(pools) != 1 {
		test.Fatal("Expected there to only be one pool")
	}

	pool2Id, err := connection.Pools.Create.CreateEc2Pool(name+"2", accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		test.Fatal(err)
	}

	pool2, err := connection.Pools.Read.GetEc2Pool(pool2Id)
	if err != nil {
		test.Fatal(err)
	}

	if pool2.Id != pool2Id {
		test.Fatal("pool2.Id mismatch")
	}

	pools, err = connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	if len(pools) != 2 {
		test.Fatal("Expected there to be two pools")
	}

	err = connection.Pools.Delete.DeleteEc2Pool(pool1Id)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-poolDeletedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear ec2 pool deletion event")
	}

	if poolDeletedEventId != pool1Id {
		test.Fatal("Bad poolId in ec2 pool deletion event")
	}

	err = connection.Pools.Delete.DeleteEc2Pool(pool1Id)
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
	PopulateDatabase()

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

	poolId, err := connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Pools.Update.SetEc2Settings(poolId, accessKey2, secretKey2, username2, baseAmiId2, securityGroupId2, vpcSubnetId2,
		instanceType2, numReadyInstances2, numMaxInstances2, rootDriveSize2, userData2)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-poolEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear ec2 pool settings updated updated event")
	}

	if poolEventId != poolId {
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

	pool, err := connection.Pools.Read.GetEc2Pool(poolId)
	if err != nil {
		test.Fatal(err)
	}

	if pool.AccessKey != accessKey2 {
		test.Fatal("pool.AccessKey mismatch")
	} else if pool.SecretKey != secretKey2 {
		test.Fatal("pool.SecretKey mismatch")
	} else if pool.Username != username2 {
		test.Fatal("pool.Username mismatch")
	} else if pool.BaseAmiId != baseAmiId2 {
		test.Fatal("pool.BaseAmiId mismatch")
	} else if pool.SecurityGroupId != securityGroupId2 {
		test.Fatal("pool.SecurityGroupId mismatch")
	} else if pool.VpcSubnetId != vpcSubnetId2 {
		test.Fatal("pool.VpcSubnetId mismatch")
	} else if pool.InstanceType != instanceType2 {
		test.Fatal("pool.InstanceType mismatch")
	} else if pool.NumReadyInstances != numReadyInstances2 {
		test.Fatal("pool.NumReadyInstances mismatch")
	} else if pool.NumMaxInstances != numMaxInstances2 {
		test.Fatal("pool.NumMaxInstances mismatch")
	} else if pool.RootDriveSize != rootDriveSize2 {
		test.Fatal("pool.RootDriveSize mismatch")
	} else if pool.UserData != userData2 {
		test.Fatal("pool.UserData mismatch")
	}

	err = connection.Pools.Update.SetEc2Settings(0, accessKey2, secretKey2, username2, baseAmiId2, securityGroupId2, vpcSubnetId2,
		instanceType2, numReadyInstances2, numMaxInstances2, rootDriveSize2, userData2)
	if _, ok := err.(resources.NoSuchPoolError); !ok {
		test.Fatal("Expected NoSuchPoolError when trying to update nonexistent ec2 pool")
	}
}
