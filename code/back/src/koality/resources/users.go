package resources

import (
	"time"
)

type User struct {
	Id           uint64
	Email        string
	FirstName    string
	LastName     string
	PasswordHash []byte
	PasswordSalt []byte
	GitHubOauth  string
	IsAdmin      bool
	Created      *time.Time
}

type SshKey struct {
	Id        uint64
	Name      string
	PublicKey string
	Created   *time.Time
}

type UsersHandler struct {
	Create       UsersCreateHandler
	Read         UsersReadHandler
	Update       UsersUpdateHandler
	Delete       UsersDeleteHandler
	Subscription UsersSubscriptionHandler
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
	AddKey(userId uint64, name, publicKey string) (uint64, error)
	RemoveKey(userId, keyId uint64) error
}

type UsersDeleteHandler interface {
	Delete(userId uint64) error
}

type UserCreatedHandler func(userId uint64)
type UserDeletedHandler func(userId uint64)
type UserNameUpdatedHandler func(userId uint64, firstName, lastName string)
type UserAdminUpdatedHandler func(userId uint64, admin bool)
type UserSshKeyAddedHandler func(userId, sshKeyId uint64)
type UserSshKeyRemovedHandler func(userId uint64, sshKeyId uint64)

type UsersSubscriptionHandler interface {
	SubscribeToCreatedEvents(updateHandler UserCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToDeletedEvents(updateHandler UserDeletedHandler) (SubscriptionId, error)
	UnsubscribeFromDeletedEvents(subscriptionId SubscriptionId) error

	SubscribeToNameUpdatedEvents(updateHandler UserNameUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromNameUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToAdminUpdatedEvents(updateHandler UserAdminUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromAdminUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToSshKeyAddedEvents(updateHandler UserSshKeyAddedHandler) (SubscriptionId, error)
	UnsubscribeFromSshKeyAddedEvents(subscriptionId SubscriptionId) error

	SubscribeToSshKeyRemovedEvents(updateHandler UserSshKeyRemovedHandler) (SubscriptionId, error)
	UnsubscribeFromSshKeyRemovedEvents(subscriptionId SubscriptionId) error
}

type InternalUsersSubscriptionHandler interface {
	FireCreatedEvent(userId uint64)
	FireDeletedEvent(userId uint64)
	FireNameUpdatedEvent(userId uint64, firstName, lastName string)
	FireAdminUpdatedEvent(userId uint64, admin bool)
	FireSshKeyAddedEvent(userId, sshKeyId uint64)
	FireSshKeyRemovedEvent(userId, sshKeyId uint64)
	UsersSubscriptionHandler
}

type UserAlreadyExistsError struct {
	Message string
}

func (err UserAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchUserError struct {
	Message string
}

func (err NoSuchUserError) Error() string {
	return err.Message
}

type KeyAlreadyExistsError struct {
	Message string
}

func (err KeyAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchKeyError struct {
	Message string
}

func (err NoSuchKeyError) Error() string {
	return err.Message
}
