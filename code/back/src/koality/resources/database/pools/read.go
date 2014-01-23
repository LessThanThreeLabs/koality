package pools

import (
	"database/sql"
	"koality/resources"
)

type Scannable interface {
	Scan(dest ...interface{}) error
}

type ReadHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalPoolsSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalPoolsSubscriptionHandler) (resources.PoolsReadHandler, error) {
	return &ReadHandler{database, verifier, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) scanEc2Pool(scannable Scannable) (*resources.Ec2Pool, error) {
	ec2Pool := new(resources.Ec2Pool)

	var baseAmiId, securityGroupId, vpcSubnetId, userData sql.NullString
	var deletedId uint64
	err := scannable.Scan(&ec2Pool.Id, &ec2Pool.Name, &ec2Pool.AccessKey, &ec2Pool.SecretKey,
		&ec2Pool.Username, &baseAmiId, &securityGroupId, &vpcSubnetId,
		&ec2Pool.InstanceType, &ec2Pool.NumReadyInstances, &ec2Pool.NumMaxInstances,
		&ec2Pool.RootDriveSize, &userData, &ec2Pool.Created, &deletedId)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchPoolError{"Unable to find ec2 pool"}
	} else if err != nil {
		return nil, err
	}

	ec2Pool.IsDeleted = ec2Pool.Id == deletedId

	if baseAmiId.Valid {
		ec2Pool.BaseAmiId = baseAmiId.String
	}
	if securityGroupId.Valid {
		ec2Pool.SecurityGroupId = securityGroupId.String
	}
	if vpcSubnetId.Valid {
		ec2Pool.VpcSubnetId = vpcSubnetId.String
	}
	if userData.Valid {
		ec2Pool.UserData = userData.String
	}
	return ec2Pool, nil
}

func (readHandler *ReadHandler) GetEc2Pool(ec2PoolId uint64) (*resources.Ec2Pool, error) {
	query := "SELECT id, name, access_key, secret_key, username, base_ami_id, security_group_id, vpc_subnet_id," +
		" instance_type, num_ready_instances, num_max_instances, root_drive_size, user_data, created, deleted" +
		" FROM ec2_pools WHERE id=$1"
	row := readHandler.database.QueryRow(query, ec2PoolId)
	return readHandler.scanEc2Pool(row)
}

func (readHandler *ReadHandler) GetAllEc2Pools() ([]resources.Ec2Pool, error) {
	query := "SELECT id, name, access_key, secret_key, username, base_ami_id, security_group_id, vpc_subnet_id," +
		" instance_type, num_ready_instances, num_max_instances, root_drive_size, user_data, created, deleted" +
		" FROM ec2_pools WHERE id != deleted"
	rows, err := readHandler.database.Query(query)
	if err != nil {
		return nil, err
	}

	ec2Pools := make([]resources.Ec2Pool, 0, 1)
	for rows.Next() {
		ec2Pool, err := readHandler.scanEc2Pool(rows)
		if err != nil {
			return nil, err
		}
		ec2Pools = append(ec2Pools, *ec2Pool)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ec2Pools, nil
}
