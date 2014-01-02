package repositories

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	createdSubscriptionManager           resources.SubscriptionManager
	deletedSubscriptionManager           resources.SubscriptionManager
	statusUpdatedSubscriptionManager     resources.SubscriptionManager
	gitHubHookUpdatedSubscriptionManager resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalRepositoriesSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToCreatedEvents(updateHandler resources.RepositoryCreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.createdSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.createdSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCreatedEvent(repository *resources.Repository) {
	subscriptionHandler.createdSubscriptionManager.Fire(repository)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToDeletedEvents(updateHandler resources.RepositoryDeletedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.deletedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromDeletedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.deletedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireDeletedEvent(repositoryId uint64) {
	subscriptionHandler.deletedSubscriptionManager.Fire(repositoryId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToStatusUpdatedEvents(updateHandler resources.RepositoryStatusUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.statusUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromStatusUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.statusUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireStatusUpdatedEvent(repositoryId uint64, status string) {
	subscriptionHandler.statusUpdatedSubscriptionManager.Fire(repositoryId, status)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToGitHubHookUpdatedEvents(updateHandler resources.RepositoryGitHubHookUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.gitHubHookUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromGitHubHookUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.gitHubHookUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireGitHubHookUpdatedEvent(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string) {
	subscriptionHandler.gitHubHookUpdatedSubscriptionManager.Fire(repositoryId, hookId, hookSecret, hookTypes)
}
