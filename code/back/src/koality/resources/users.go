package resources

import (
	"time"
)

type User struct {
	Id           int
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	PasswordSalt string
	GitHubOauth  string
	Admin        bool
	Created      *time.Time
	Deleted      *time.Time
}

type SshKey struct {
	Alias     string
	PublicKey string
	Created   *time.Time
}

type UsersHandler struct {
	Read UsersReadHandler
	// Update UsersUpdateHandler
}

type UsersReadHandler interface {
	Get(userId int) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]User, error)
	GetKeys(userId int) ([]SshKey, error)
}

type UsersUpdateHandler interface {
	// SetName(userId int, firstName, lastName string) error
}

type NoSuchUserError error
