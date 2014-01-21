package resources

import (
	"time"
)

type Repository struct {
	Id        uint64
	Name      string
	Status    string
	VcsType   string
	LocalUri  string
	RemoteUri string
	Created   *time.Time
	GitHub    *GitHubMetadata
	IsDeleted bool
}

type GitHubMetadata struct {
	Owner      string
	Name       string
	HookId     int64
	HookSecret string
	HookTypes  []string
}

type RepositoriesHandler struct {
	Create       RepositoriesCreateHandler
	Read         RepositoriesReadHandler
	Update       RepositoriesUpdateHandler
	Delete       RepositoriesDeleteHandler
	Subscription RepositoriesSubscriptionHandler
}

type RepositoriesCreateHandler interface {
	Create(name, vcsType, localUri, remoteUri string) (*Repository, error)
	CreateWithGitHub(name, vcsType, localUri, remoteUri, gitHubOwner, gitHubName string) (*Repository, error)
}

type RepositoriesReadHandler interface {
	Get(repositoryId uint64) (*Repository, error)
	GetByLocalUri(localUri string) (*Repository, error)
	GetAll() ([]Repository, error)
}

type RepositoriesUpdateHandler interface {
	SetStatus(repositoryId uint64, status string) error
	SetGitHubHook(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string) error
	ClearGitHubHook(repositoryId uint64) error
}

type RepositoriesDeleteHandler interface {
	Delete(repositoryId uint64) error
}

type RepositoryCreatedHandler func(repository *Repository)
type RepositoryDeletedHandler func(repositoryId uint64)
type RepositoryStatusUpdatedHandler func(repositoryId uint64, status string)
type RepositoryGitHubHookUpdatedHandler func(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string)

type RepositoriesSubscriptionHandler interface {
	SubscribeToCreatedEvents(updateHandler RepositoryCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToDeletedEvents(updateHandler RepositoryDeletedHandler) (SubscriptionId, error)
	UnsubscribeFromDeletedEvents(subscriptionId SubscriptionId) error

	SubscribeToStatusUpdatedEvents(updateHandler RepositoryStatusUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStatusUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToGitHubHookUpdatedEvents(updateHandler RepositoryGitHubHookUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromGitHubHookUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalRepositoriesSubscriptionHandler interface {
	FireCreatedEvent(repository *Repository)
	FireDeletedEvent(repositoryId uint64)
	FireStatusUpdatedEvent(repositoryId uint64, status string)
	FireGitHubHookUpdatedEvent(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string)
	RepositoriesSubscriptionHandler
}
