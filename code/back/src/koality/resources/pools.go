package resources

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
