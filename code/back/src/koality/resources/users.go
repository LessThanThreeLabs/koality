package resources

import (
	"time"
)

type User struct {
	Id           int64
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	PasswordSalt string
	GitHubOauth  string
	IsAdmin      bool
	Created      *time.Time
}

type SshKey struct {
	Id        int64
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
	Create(email, firstName, lastName, passwordHash, passwordSalt string, admin bool) (int64, error)
}

type UsersReadHandler interface {
	Get(userId int64) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]User, error)
	GetKeys(userId int64) ([]SshKey, error)
}

type UsersUpdateHandler interface {
	SetName(userId int64, firstName, lastName string) error
	SetPassword(userId int64, passwordHash, passwordSalt string) error
	SetGitHubOauth(userId int64, gitHubOauth string) error
	SetAdmin(userId int64, admin bool) error
	AddKey(userId int64, alias, publicKey string) (int64, error)
	RemoveKey(userId, keyId int64) error
}

type UsersDeleteHandler interface {
	Delete(userId int64) error
}

type UserAlreadyExistsError error
type NoSuchUserError error

type KeyAlreadyExistsError error
type NoSuchKeyError error
