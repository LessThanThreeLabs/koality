package database

import (

	// "koality/resources"
	"testing"
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

	poolId, err := connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
		instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		test.Fatal(err)
	}

	pool, err := connection.Pools.Read.GetEc2Pool(poolId)
	if err != nil {
		test.Fatal(err)
	}

	if pool.Id != poolId {
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
}
