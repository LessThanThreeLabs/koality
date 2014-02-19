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
	eventsHandler.subscriptionsRWMutex.Lock()
	defer eventsHandler.subscriptionsRWMutex.Unlock()

	eventsHandler.subscriptions[subscriptionType] = append(eventsHandler.subscriptions[subscriptionType], subscription)
	return nil
}

func (eventsHandler *EventsHandler) removeFromSubscriptions(subscriptionType subscriptionType, userId, subscriptionId uint64) error {
	eventsHandler.subscriptionsRWMutex.Lock()
	defer eventsHandler.subscriptionsRWMutex.Unlock()

	subscriptionIndex := eventsHandler.getSubscriptionIndex(subscriptionType, userId, subscriptionId)
	if subscriptionIndex < 0 {
		return fmt.Errorf("Unable to find subscription %d for user %d", subscriptionId, userId)
	}

	eventsHandler.subscriptions[subscriptionType] = append(
		eventsHandler.subscriptions[subscriptionType][:subscriptionIndex],
		eventsHandler.subscriptions[subscriptionType][subscriptionIndex+1:]...)
	return nil
}

func (eventsHandler *EventsHandler) handleEvent(subscriptionType subscriptionType, resourceId uint64, data interface{}) {
	subscriptionIdsToDelete := make([]uint64, 0)

	eventsHandler.subscriptionsRWMutex.RLock()
	for _, subscription := range eventsHandler.subscriptions[subscriptionType] {
		if subscription.allResources || subscription.resourceId == resourceId {
			message := eventData{subscription.id, data}
			if err := eventsHandler.websocketsManager.SendJson(subscription.websocketId, message); err != nil {
				subscriptionIdsToDelete = append(subscriptionIdsToDelete, subscription.id)
			}
		}
	}
	eventsHandler.subscriptionsRWMutex.RUnlock()

	eventsHandler.removeSubscriptions(subscriptionType, subscriptionIdsToDelete)
}

func (eventsHandler *EventsHandler) removeSubscriptions(subscriptionType subscriptionType, subscriptionIdsToDelete []uint64) {
	eventsHandler.subscriptionsRWMutex.Lock()
	defer eventsHandler.subscriptionsRWMutex.Unlock()

	subscriptionsToRetain := make([]subscription, 0, len(eventsHandler.subscriptions[subscriptionType]))

loop:
	for _, subscription := range eventsHandler.subscriptions[subscriptionType] {
		for _, subscriptionIdToDelete := range subscriptionIdsToDelete {
			if subscription.id == subscriptionIdToDelete {
				continue loop
			}
		}
		subscriptionsToRetain = append(subscriptionsToRetain, subscription)
	}

	eventsHandler.subscriptions[subscriptionType] = subscriptionsToRetain
}
