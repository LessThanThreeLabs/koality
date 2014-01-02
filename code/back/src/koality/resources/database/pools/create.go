package pools

import (
	"database/sql"
	"koality/resources"
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	readHandler         resources.PoolsReadHandler
	subscriptionHandler resources.InternalPoolsSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.PoolsReadHandler,
	subscriptionHandler resources.InternalPoolsSubscriptionHandler) (resources.PoolsCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
	numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) (*resources.Ec2Pool, error) {

	err := createHandler.getEc2ParamsError(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType,
		numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO ec2_pools (name, access_key, secret_key, username," +
		" base_ami_id, security_group_id, vpc_subnet_id, instance_type," +
		" num_ready_instances, num_max_instances, root_drive_size, user_data)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id"
	err = createHandler.database.QueryRow(query, name, accessKey, secretKey, username,
		baseAmiId, securityGroupId, vpcSubnetId, instanceType,
		numReadyInstances, numMaxInstances, rootDriveSize, userData).Scan(&id)
	if err != nil {
		return nil, err
	}

	ec2Pool, err := createHandler.readHandler.GetEc2Pool(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireEc2CreatedEvent(ec2Pool)
	return ec2Pool, nil
}

func (createHandler *CreateHandler) getEc2ParamsError(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
	numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) error {

	if err := createHandler.verifier.verifyName(name); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyEc2AccessKey(accessKey); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyEc2SecretKey(secretKey); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyUsername(username); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyEc2BaseAmiId(baseAmiId); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyEc2SecurityGroupId(securityGroupId); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyEc2VpcSubnetId(vpcSubnetId); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyEc2InstanceType(instanceType); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyReadyAndMaxInstances(numReadyInstances, numMaxInstances); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyEc2RootDriveSize(rootDriveSize); err != nil {
		return err
	}
	return nil
}
