package database

import (
	"koality/resources"
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

func TestUsersEc2Settings(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
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
}

func TestDeleteEc2Pool(test *testing.T) {
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

	err = connection.Pools.Delete.DeleteEc2Pool(poolId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Pools.Delete.DeleteEc2Pool(poolId)
	if _, ok := err.(resources.NoSuchPoolError); !ok {
		test.Fatal("Expected NoSuchPoolError when trying to delete same pool twice")
	}

	pools, err := connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	if len(pools) != 0 {
		test.Fatal("Expected there to only no pools")
	}
}
