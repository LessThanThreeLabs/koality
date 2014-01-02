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

func NewInternalSubscriptionHandler() (resources.InternalUsersSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToCreatedEvents(updateHandler resources.UserCreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.createdSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.createdSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCreatedEvent(user *resources.User) {
	subscriptionHandler.createdSubscriptionManager.Fire(user)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToDeletedEvents(updateHandler resources.UserDeletedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.deletedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromDeletedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.deletedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireDeletedEvent(userId uint64) {
	subscriptionHandler.deletedSubscriptionManager.Fire(userId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToNameUpdatedEvents(updateHandler resources.UserNameUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.nameUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromNameUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.nameUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireNameUpdatedEvent(userId uint64, firstName, lastName string) {
	subscriptionHandler.nameUpdatedSubscriptionManager.Fire(userId, firstName, lastName)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToAdminUpdatedEvents(updateHandler resources.UserAdminUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.adminUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromAdminUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.adminUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireAdminUpdatedEvent(userId uint64, admin bool) {
	subscriptionHandler.adminUpdatedSubscriptionManager.Fire(userId, admin)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSshKeyAddedEvents(updateHandler resources.UserSshKeyAddedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.sshKeyAddedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSshKeyAddedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.sshKeyAddedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireSshKeyAddedEvent(userId, sshKeyId uint64) {
	subscriptionHandler.sshKeyAddedSubscriptionManager.Fire(userId, sshKeyId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSshKeyRemovedEvents(updateHandler resources.UserSshKeyRemovedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.sshKeyRemovedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSshKeyRemovedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.sshKeyRemovedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireSshKeyRemovedEvent(userId, sshKeyId uint64) {
	subscriptionHandler.sshKeyRemovedSubscriptionManager.Fire(userId, sshKeyId)
}
