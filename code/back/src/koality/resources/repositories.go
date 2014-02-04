package resources

import (
	"time"
)

type Repository struct {
	Id        uint64
	Name      string
	Status    string
	VcsType   string
	RemoteUri string
	Created   *time.Time
	GitHub    *RepositoryGitHubMetadata
	IsDeleted bool
}

type RepositoryGitHubMetadata struct {
	Owner      string
	Name       string
	HookId     int64
	HookSecret string
	HookTypes  []string
	OAuthToken string
}

type RepositoriesHandler struct {
	Create       RepositoriesCreateHandler
	Read         RepositoriesReadHandler
	Update       RepositoriesUpdateHandler
	Delete       RepositoriesDeleteHandler
	Subscription RepositoriesSubscriptionHandler
}

type RepositoriesCreateHandler interface {
	Create(name, vcsType, remoteUri string) (*Repository, error)
	CreateWithGitHub(name, remoteUri, gitHubOwner, gitHubName, oAuthToken string) (*Repository, error)
}

type RepositoriesReadHandler interface {
	Get(repositoryId uint64) (*Repository, error)
	GetByName(name string) (*Repository, error)
	GetByGitHubInfo(ownerName, repositoryName string) (*Repository, error)
	GetAll() ([]Repository, error)
}

type RepositoriesUpdateHandler interface {
	SetStatus(repositoryId uint64, status string) error
	SetGitHubOAuthToken(repositoryId uint64, oAuthToken string) error
	ClearGitHubOAuthToken(repositoryId uint64) error
	SetGitHubHook(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string) error
	ClearGitHubHook(repositoryId uint64) error
}

type RepositoriesDeleteHandler interface {
	Delete(repositoryId uint64) error
}

type RepositoryCreatedHandler func(repository *Repository)
type RepositoryDeletedHandler func(repositoryId uint64)
type RepositoryStatusUpdatedHandler func(repositoryId uint64, status string)
type RepositoryGitHubOAuthTokenUpdatedHandler func(repositoryId uint64, oAuthToken string)
type RepositoryGitHubOAuthTokenClearedHandler func(repositoryId uint64)
type RepositoryGitHubHookUpdatedHandler func(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string)
type RepositoryGitHubHookClearedHandler func(repositoryId uint64)

type RepositoriesSubscriptionHandler interface {
	SubscribeToCreatedEvents(updateHandler RepositoryCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToDeletedEvents(updateHandler RepositoryDeletedHandler) (SubscriptionId, error)
	UnsubscribeFromDeletedEvents(subscriptionId SubscriptionId) error

	SubscribeToStatusUpdatedEvents(updateHandler RepositoryStatusUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStatusUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToGitHubOAuthTokenUpdatedEvents(updateHandler RepositoryGitHubOAuthTokenUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromGitHubOAuthTokenUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToGitHubOAuthTokenClearedEvents(updateHandler RepositoryGitHubOAuthTokenClearedHandler) (SubscriptionId, error)
	UnsubscribeFromGitHubOAuthTokenClearedEvents(subscriptionId SubscriptionId) error

	SubscribeToGitHubHookUpdatedEvents(updateHandler RepositoryGitHubHookUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromGitHubHookUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToGitHubHookClearedEvents(updateHandler RepositoryGitHubHookClearedHandler) (SubscriptionId, error)
	UnsubscribeFromGitHubHookClearedEvents(subscriptionId SubscriptionId) error
}

type InternalRepositoriesSubscriptionHandler interface {
	FireCreatedEvent(repository *Repository)
	FireDeletedEvent(repositoryId uint64)
	FireStatusUpdatedEvent(repositoryId uint64, status string)
	FireGitHubOAuthTokenUpdatedEvent(repositoryId uint64, oAuthToken string)
	FireGitHubOAuthTokenClearedEvent(repositoryId uint64)
	FireGitHubHookUpdatedEvent(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string)
	FireGitHubHookClearedEvent(repositoryId uint64)
	RepositoriesSubscriptionHandler
}
