package resources

import (
	"time"
)

type User struct {
	Id           uint64
	Email        string
	FirstName    string
	LastName     string
	PasswordHash *[]byte
	PasswordSalt *[]byte
	GitHubOauth  string
	IsAdmin      bool
	Created      *time.Time
}

type SshKey struct {
	Id        uint64
	Alias     string
	PublicKey string
	Created   *time.Time
}

type UsersHandler struct {
	Create UsersCreateHandler
	Read   UsersReadHandler
	Update UsersUpdateHandler
	Delete UsersDeleteHandler
}

type UsersCreateHandler interface {
	Create(email, firstName, lastName string, passwordHash, passwordSalt []byte, admin bool) (uint64, error)
}

type UsersReadHandler interface {
	Get(userId uint64) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]User, error)
	GetKeys(userId uint64) ([]SshKey, error)
}

type UsersUpdateHandler interface {
	SetName(userId uint64, firstName, lastName string) error
	SetPassword(userId uint64, passwordHash, passwordSalt []byte) error
	SetGitHubOauth(userId uint64, gitHubOauth string) error
	SetAdmin(userId uint64, admin bool) error
	AddKey(userId uint64, alias, publicKey string) (uint64, error)
	RemoveKey(userId, keyId uint64) error
}

type UsersDeleteHandler interface {
	Delete(userId uint64) error
}

type UserAlreadyExistsError error
type NoSuchUserError error

type KeyAlreadyExistsError error
type NoSuchKeyError error
