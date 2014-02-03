package resources

import (
	"time"
)

type DebugInstance struct {
	Id             uint64
	PoolId         uint64
	VerificationId uint64
	InstanceId     string
	Expires        *time.Time
}

type DebugInstancesHandler struct {
	Create       DebugInstancesCreateHandler
	Read         DebugInstancesReadHandler
	Subscription DebugInstancesSubscriptionHandler
}

type DebugInstancesCreateHandler interface {
	Create(poolId uint64, instanceId string, expiration *time.Time, verification *CoreVerificationInformation) (*DebugInstance, error)
}

type DebugInstancesReadHandler interface {
	Get(debugInstanceId uint64) (*DebugInstance, error)
	GetAllRunning() ([]DebugInstance, error)
}

type DebugInstanceCreatedHandler func(debugInstance *DebugInstance)

type DebugInstancesSubscriptionHandler interface {
	SubscribeToCreatedEvents(createHandler DebugInstanceCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error
}

type InternalDebugInstancesSubscriptionHandler interface {
	FireCreatedEvent(debugInstance *DebugInstance)
	DebugInstancesSubscriptionHandler
}
