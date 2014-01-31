package debugInstances

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	createdSubscriptionManager resources.SubscriptionManager
	deletedSubscriptionManager resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalDebugInstancesSubscriptionHandler, error) {
	return new(SubscriptionHandler), nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToDebugInstanceCreatedEvents(createHandler resources.DebugInstanceCreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.createdSubscriptionManager.Add(createHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.createdSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCreatedEvent(debugInstance *resources.DebugInstance) {
	subscriptionHandler.createdSubscriptionManager.Fire(debugInstance)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToDebugInstanceDeletedEvents(deleteHandler resources.DebugInstanceDeletedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.deletedSubscriptionManager.Add(deleteHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromDebugInstanceDeletedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.deletedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireDeletedEvent(debugInstanceId uint64) {
	subscriptionHandler.deletedSubscriptionManager.Fire(debugInstanceId)
}
