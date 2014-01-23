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
	IsDeleted         bool
}

type PoolsHandler struct {
	Create       PoolsCreateHandler
	Read         PoolsReadHandler
	Update       PoolsUpdateHandler
	Delete       PoolsDeleteHandler
	Subscription PoolsSubscriptionHandler
}

type PoolsCreateHandler interface {
	CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
		numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) (*Ec2Pool, error)
}

type PoolsReadHandler interface {
	GetEc2Pool(ec2PoolId uint64) (*Ec2Pool, error)
	GetAllEc2Pools() ([]Ec2Pool, error)
}

type PoolsUpdateHandler interface {
	SetEc2Settings(poolId uint64, accessKey, secretKey,
		username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
		numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) error
}

type PoolsDeleteHandler interface {
	DeleteEc2Pool(poolId uint64) error
}

type PoolEc2CreatedHandler func(ec2Pool *Ec2Pool)
type PoolEc2DeletedHandler func(ec2PoolId uint64)
type PoolEc2SettingsUpdatedHandler func(ec2PoolId uint64, accessKey, secretKey,
	username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
	numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string)

type PoolsSubscriptionHandler interface {
	SubscribeToEc2CreatedEvents(updateHandler PoolEc2CreatedHandler) (SubscriptionId, error)
	UnsubscribeFromEc2CreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToEc2DeletedEvents(updateHandler PoolEc2DeletedHandler) (SubscriptionId, error)
	UnsubscribeFromEc2DeletedEvents(subscriptionId SubscriptionId) error

	SubscribeToEc2SettingsUpdatedEvents(updateHandler PoolEc2SettingsUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromEc2SettingsUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalPoolsSubscriptionHandler interface {
	FireEc2CreatedEvent(ec2Pool *Ec2Pool)
	FireEc2DeletedEvent(ec2PoolId uint64)
	FireEc2SettingsUpdatedEvent(ec2PoolId uint64, accessKey, secretKey,
		username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
		numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string)
	PoolsSubscriptionHandler
}
