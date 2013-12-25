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
}

type GitHubMetadata struct {
	Owner      string
	Name       string
	HookId     int64
	HookSecret string
	HookTypes  []string
}

type RepositoriesHandler struct {
	Create RepositoriesCreateHandler
	Read   RepositoriesReadHandler
	Update RepositoriesUpdateHandler
	Delete RepositoriesDeleteHandler
}

type RepositoriesCreateHandler interface {
	Create(name, vcsType, localUri, remoteUri string) (uint64, error)
	CreateWithGitHub(name, vcsType, localUri, remoteUri, gitHubOwner, gitHubName string) (uint64, error)
}

type RepositoriesReadHandler interface {
	Get(repositoryId uint64) (*Repository, error)
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

type RepositoryAlreadyExistsError struct {
	Message string
}

func (err RepositoryAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchRepositoryError struct {
	Message string
}

func (err NoSuchRepositoryError) Error() string {
	return err.Message
}

type InvalidRepositoryStatusError struct {
	Message string
}

func (err InvalidRepositoryStatusError) Error() string {
	return err.Message
}

type NoSuchRepositoryHookError struct {
	Message string
}

func (err NoSuchRepositoryHookError) Error() string {
	return err.Message
}
