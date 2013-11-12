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

type UsersHandler struct {
	Read UsersReadHandler
	// Update UsersUpdateHandler
}

type UsersReadHandler interface {
	Get(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]*User, error)
}

// type UsersUpdateHandler interface {
// 	SetName(id int, firstName, lastName string) error
// }

type NoSuchUserError error
