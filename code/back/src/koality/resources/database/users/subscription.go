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
	return subscriptionHandler.createdSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromUserCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.createdSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) fireCreatedEvent(userId uint64, firstName, lastName string, admin bool) {
	subscriptionHandler.createdSubscriptionManager.Fire(userId, firstName, lastName, admin)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToUserDeletedEvents(updateHandler resources.UserDeletedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.deletedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromUserDeletedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.deletedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) fireDeletedEvent(userId uint64) {
	subscriptionHandler.deletedSubscriptionManager.Fire(userId)
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
	return subscriptionHandler.adminUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromAdminEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.adminUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) fireAdminEvent(userId uint64, admin bool) {
	subscriptionHandler.adminUpdatedSubscriptionManager.Fire(userId, admin)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSshKeyAddedEvents(updateHandler resources.UserSshKeyAddedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.sshKeyAddedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSshKeyAddedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.sshKeyAddedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) fireSshKeyAddedEvent(userId, sshKeyId uint64, name string) {
	subscriptionHandler.sshKeyAddedSubscriptionManager.Fire(userId, sshKeyId, name)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSshKeyRemovedEvents(updateHandler resources.UserSshKeyRemovedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.sshKeyRemovedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSshKeyRemovedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.sshKeyRemovedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) fireSshKeyRemovedEvent(userId, sshKeyId uint64) {
	subscriptionHandler.sshKeyRemovedSubscriptionManager.Fire(userId, sshKeyId)
}
