package events

import (
	"fmt"
)

const (
	userCreatedSubscriptions = iota
	userDeletedSubscriptions
	userNameUpdatedSubscriptions
	userAdminUpdatedSubscriptions
	userSshKeyAddedSubscriptions
	userSshKeyRemovedSubscriptions
)

type subscriptionType int

type subscription struct {
	id           uint64
	userId       uint64
	websocketId  string
	allResources bool
	resourceId   uint64 // 0 if allResources == true
}

func (eventsHandler *EventsHandler) getSubscriptionIndex(subscriptionType subscriptionType, userId, subscriptionId uint64) int {
	for index, subscription := range eventsHandler.subscriptions[subscriptionType] {
		if subscription.id == subscriptionId && subscription.userId == userId {
			return index
		}
	}
	return -1
}

func (eventsHandler *EventsHandler) addToSubscriptions(subscriptionType subscriptionType, subscription subscription) error {
	eventsHandler.subscriptionsMutex.Lock()
	defer eventsHandler.subscriptionsMutex.Unlock()

	eventsHandler.subscriptions[subscriptionType] = append(eventsHandler.subscriptions[subscriptionType], subscription)
	return nil
}

func (eventsHandler *EventsHandler) removeFromSubscriptions(subscriptionType subscriptionType, userId, subscriptionId uint64) error {
	eventsHandler.subscriptionsMutex.Lock()
	defer eventsHandler.subscriptionsMutex.Unlock()

	subscriptionIndex := eventsHandler.getSubscriptionIndex(subscriptionType, userId, subscriptionId)
	if subscriptionIndex < 0 {
		return fmt.Errorf("Unable to find subscription %d for user %d", subscriptionId, userId)
	}

	eventsHandler.subscriptions[subscriptionType] = append(
		eventsHandler.subscriptions[subscriptionType][:subscriptionIndex],
		eventsHandler.subscriptions[subscriptionType][subscriptionIndex+1:]...)
	return nil
}
