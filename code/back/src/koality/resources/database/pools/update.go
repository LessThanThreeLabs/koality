package pools

import (
	"database/sql"
	"koality/resources"
)

type UpdateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier) (resources.PoolsUpdateHandler, error) {
	return &UpdateHandler{database, verifier}, nil
}

func (updateHandler *UpdateHandler) updatePool(query string, params ...interface{}) error {
	result, err := updateHandler.database.Exec(query, params...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchUserError{"Unable to find pool"}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetEc2Settings(poolId uint64, accessKey, secretKey,
	username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
	numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) error {

	err := updateHandler.getEc2ParamsError(accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType,
		numReadyInstances, numMaxInstances, rootDriveSize, userData)
	if err != nil {
		return err
	}

	query := "UPDATE ec2_pools SET access_key=$1, secret_key=$2, username=$3," +
		" base_ami_id=$4, security_group_id=$5, vpc_subnet_id=$6, instance_type=$7," +
		" num_ready_instances=$8, num_max_instances=$9, root_drive_size=$10, user_data=$11" +
		" WHERE id=$12"
	err = updateHandler.updatePool(query, accessKey, secretKey, username,
		baseAmiId, securityGroupId, vpcSubnetId, instanceType,
		numReadyInstances, numMaxInstances, rootDriveSize, userData, poolId)
	return err
}

func (UpdateHandler *UpdateHandler) getEc2ParamsError(accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
	numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) error {

	if err := UpdateHandler.verifier.verifyEc2AccessKey(accessKey); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyEc2SecretKey(secretKey); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyUsername(username); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyEc2BaseAmiId(baseAmiId); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyEc2SecurityGroupId(securityGroupId); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyEc2VpcSubnetId(vpcSubnetId); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyEc2InstanceType(instanceType); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyReadyAndMaxInstances(numReadyInstances, numMaxInstances); err != nil {
		return err
	}
	if err := UpdateHandler.verifier.verifyEc2RootDriveSize(rootDriveSize); err != nil {
		return err
	}
	return nil
}
