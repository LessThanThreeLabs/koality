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
	IsDeleted      bool
}

type DebugInstancesHandler struct {
	Create       DebugInstancesCreateHandler
	Read         DebugInstancesReadHandler
	Delete       DebugInstancesDeleteHandler
	Subscription DebugInstancesSubscriptionHandler
}

type DebugInstancesCreateHandler interface {
	Create(poolId uint64, instanceId string, expiration *time.Time, verification *CoreVerificationInformation) (*DebugInstance, error)
}

type DebugInstancesReadHandler interface {
	Get(debugInstanceId uint64) (*DebugInstance, error)
	GetAllRunning() ([]DebugInstance, error)
}

type DebugInstancesDeleteHandler interface {
	Delete(debugInstanceId uint64) error
}

type DebugInstanceCreatedHandler func(debugInstance *DebugInstance)
type DebugInstanceDeletedHandler func(debugInstanceId uint64)

type DebugInstancesSubscriptionHandler interface {
	SubscribeToCreatedEvents(createHandler DebugInstanceCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToDeletedEvents(deletedHandler DebugInstanceDeletedHandler) (SubscriptionId, error)
	UnsubscribeFromDeletedEvents(subscriptionId SubscriptionId) error
}

type InternalDebugInstancesSubscriptionHandler interface {
	FireCreatedEvent(debugInstance *DebugInstance)
	FireDeletedEvent(debugInstanceId uint64)
	DebugInstancesSubscriptionHandler
}
