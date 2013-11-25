package resources

import (
	"time"
)

type Repository struct {
	Id        uint64
	Name      string
	VcsType   string
	LocalUri  string
	RemoteUri string
	Created   *time.Time
	GitHub    *GitHubMetadata
}

type GitHubMetadata struct {
	Owner      string
	Name       string
	HookId     int
	HookSecret string
	HookTypes  []string
}

type RepositoriesHandler struct {
	Create RepositoriesCreateHandler
	// Read   UsersReadHandler
	// Update UsersUpdateHandler
	Delete RepositoriesDeleteHandler
}

type RepositoriesCreateHandler interface {
	Create(name, vcsType, localUri, remoteUri string) (uint64, error)
}

// type RepositoriesReadHandler interface {
// 	Get(userId int64) (*User, error)
// 	GetByEmail(email string) (*User, error)
// 	GetAll() ([]User, error)
// 	GetKeys(userId int64) ([]SshKey, error)
// }

// type RepositoriesUpdateHandler interface {
// 	SetName(userId int64, firstName, lastName string) error
// 	SetPassword(userId int64, passwordHash, passwordSalt string) error
// 	SetGitHubOauth(userId int64, gitHubOauth string) error
// 	SetAdmin(userId int64, admin bool) error
// 	AddKey(userId int64, alias, publicKey string) (int64, error)
// 	RemoveKey(userId, keyId int64) error
// }

type RepositoriesDeleteHandler interface {
	Delete(repositoryId uint64) error
}

type RepositoryAlreadyExistsError error
type NoSuchRepositoryError error
