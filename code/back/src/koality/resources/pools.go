package resources

import (
	"time"
)

type Ec2Pool struct {
	Id                uint64
	Name              string
	AccessKey         string
	SecretKey         string
	Username          string
	BaseAmiId         string
	SecurityGroupId   string
	VpcSubnetId       string
	InstanceType      string
	NumReadyInstances uint64
	NumMaxInstances   uint64
	RootDriveSize     uint64
	UserData          string
	Created           *time.Time
}

type PoolsHandler struct {
	Create PoolsCreateHandler
	Read   PoolsReadHandler
	Update PoolsUpdateHandler
	Delete PoolsDeleteHandler
}

type PoolsCreateHandler interface {
	CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
		numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) (uint64, error)
}

type PoolsReadHandler interface {
	GetEc2Pool(ec2PoolId uint64) (*Ec2Pool, error)
	GetAllEc2Pools() ([]Ec2Pool, error)
}

type PoolsUpdateHandler interface {
}

type PoolsDeleteHandler interface {
}

type NoSuchPoolError struct {
	Message string
}

func (err NoSuchPoolError) Error() string {
	return err.Message
}

type PoolAlreadyExistsError struct {
	Message string
}

func (err PoolAlreadyExistsError) Error() string {
	return err.Message
}
