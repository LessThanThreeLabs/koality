package pools

import (
	"database/sql"
	"koality/resources"
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalPoolsSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalPoolsSubscriptionHandler) (resources.PoolsCreateHandler, error) {
	return &CreateHandler{database, verifier, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
	numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) (uint64, error) {

	err := createHandler.getEc2ParamsError(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType,
		numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	createHandler.subscriptionHandler.FireEc2CreatedEvent(id)
	return id, nil
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
