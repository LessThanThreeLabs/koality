package repositories

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	createdSubscriptionManager                 resources.SubscriptionManager
	deletedSubscriptionManager                 resources.SubscriptionManager
	statusUpdatedSubscriptionManager           resources.SubscriptionManager
	gitHubOAuthTokenUpdatedSubscriptionManager resources.SubscriptionManager
	gitHubOAuthTokenClearedSubscriptionManager resources.SubscriptionManager
	gitHubHookUpdatedSubscriptionManager       resources.SubscriptionManager
	gitHubHookClearedSubscriptionManager       resources.SubscriptionManager
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

func (subscriptionHandler *SubscriptionHandler) SubscribeToGitHubOAuthTokenUpdatedEvents(updateHandler resources.RepositoryGitHubOAuthTokenUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.gitHubOAuthTokenUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromGitHubOAuthTokenUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.gitHubOAuthTokenUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireGitHubOAuthTokenUpdatedEvent(repositoryId uint64, oAuthToken string) {
	subscriptionHandler.gitHubOAuthTokenUpdatedSubscriptionManager.Fire(repositoryId, oAuthToken)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToGitHubOAuthTokenClearedEvents(updateHandler resources.RepositoryGitHubOAuthTokenClearedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.gitHubOAuthTokenClearedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromGitHubOAuthTokenClearedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.gitHubOAuthTokenClearedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireGitHubOAuthTokenClearedEvent(repositoryId uint64) {
	subscriptionHandler.gitHubOAuthTokenClearedSubscriptionManager.Fire(repositoryId)
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

func (subscriptionHandler *SubscriptionHandler) SubscribeToGitHubHookClearedEvents(updateHandler resources.RepositoryGitHubHookClearedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.gitHubHookClearedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromGitHubHookClearedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.gitHubHookClearedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireGitHubHookClearedEvent(repositoryId uint64) {
	subscriptionHandler.gitHubHookClearedSubscriptionManager.Fire(repositoryId)
}
