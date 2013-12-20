package users

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	createdSubscriptionManager       resources.SubscriptionManager
	deletedSubscriptionManager       resources.SubscriptionManager
	nameUpdatedSubscriptionManager   resources.SubscriptionManager
	adminUpdatedSubscriptionManager  resources.SubscriptionManager
	sshKeyAddedSubscriptionManager   resources.SubscriptionManager
	sshKeyRemovedSubscriptionManager resources.SubscriptionManager
}

func NewSubscriptionHandler() (resources.UsersSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToUserCreatedEvents(updateHandler resources.UserCreatedHandler) (resources.SubscriptionId, error) {
	return 0, nil
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromUserCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToUserDeletedEvents(updateHandler resources.UserDeletedHandler) (resources.SubscriptionId, error) {
	return 0, nil
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromUserDeletedEvents(subscriptionId resources.SubscriptionId) error {
	return nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToNameEvents(updateHandler resources.UserNameUpdateHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.nameUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromNameEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.nameUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) fireNameEvent(userId uint64, firstName, lastName string) {
	subscriptionHandler.nameUpdatedSubscriptionManager.Fire(userId, firstName, lastName)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToAdminEvents(updateHandler resources.UserAdminUpdateHandler) (resources.SubscriptionId, error) {
	return 0, nil
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromAdminEvents(subscriptionId resources.SubscriptionId) error {
	return nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSshKeyAddedEvents(updateHandler resources.UserSshKeyAddedHandler) (resources.SubscriptionId, error) {
	return 0, nil
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSshKeyAddedEvents(subscriptionId resources.SubscriptionId) error {
	return nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSshKeyRemovedEvents(updateHandler resources.UserSshKeyRemovedHandler) (resources.SubscriptionId, error) {
	return 0, nil
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSshKeyRemovedEvents(subscriptionId resources.SubscriptionId) error {
	return nil
}
