package settings

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	repositoryKeyPairUpdatedSubscriptionManager resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalSettingsSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToRepositoryKeyPairUpdatedEvents(updateHandler resources.RepositoryKeyPairUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.repositoryKeyPairUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromRepositoryKeyPairUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.repositoryKeyPairUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireRepositoryKeyPairUpdatedEvent(keyPair *resources.RepositoryKeyPair) {
	subscriptionHandler.repositoryKeyPairUpdatedSubscriptionManager.Fire(keyPair)
}
