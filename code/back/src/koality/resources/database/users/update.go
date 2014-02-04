package users

import (
	"database/sql"
	"fmt"
	"koality/resources"
	"strings"
)

type UpdateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalUsersSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalUsersSubscriptionHandler) (resources.UsersUpdateHandler, error) {
	return &UpdateHandler{database, verifier, subscriptionHandler}, nil
}

func (updateHandler *UpdateHandler) updateUser(query string, params ...interface{}) error {
	result, err := updateHandler.database.Exec(query, params...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchUserError{"Unable to find user"}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetName(userId uint64, firstName, lastName string) error {
	if err := updateHandler.verifier.verifyFirstName(firstName); err != nil {
		return err
	} else if err := updateHandler.verifier.verifyLastName(lastName); err != nil {
		return err
	}
	query := "UPDATE users SET first_name=$1, last_name=$2 WHERE id=$3"
	err := updateHandler.updateUser(query, firstName, lastName, userId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireNameUpdatedEvent(userId, firstName, lastName)
	return nil
}

func (updateHandler *UpdateHandler) SetPassword(userId uint64, passwordHash, passwordSalt []byte) error {
	query := "UPDATE users SET password_hash=$1, password_salt=$2 WHERE id=$3"
	return updateHandler.updateUser(query, passwordHash, passwordSalt, userId)
}

func (updateHandler *UpdateHandler) SetGitHubOAuth(userId uint64, gitHubOAuth string) error {
	query := "UPDATE users SET github_oauth=$1 WHERE id=$2"
	return updateHandler.updateUser(query, gitHubOAuth, userId)
}

func (updateHandler *UpdateHandler) SetAdmin(userId uint64, admin bool) error {
	query := "UPDATE users SET is_admin=$1 WHERE id=$2"
	err := updateHandler.updateUser(query, admin, userId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireAdminUpdatedEvent(userId, admin)
	return nil
}

func (updateHandler *UpdateHandler) AddKey(userId uint64, name, publicKey string) (uint64, error) {
	publicKeyFields := strings.Fields(publicKey)
	if len(publicKeyFields) < 2 {
		return 0, fmt.Errorf("Public key must contain at least two fields")
	}
	sanitizedPublicKey := strings.Join(publicKeyFields[:2], " ")

	if err := updateHandler.verifier.verifyKeyName(userId, name); err != nil {
		return 0, err
	} else if err := updateHandler.verifier.verifyPublicKey(sanitizedPublicKey); err != nil {
		return 0, err
	} else if err := updateHandler.verifier.verifyUserExists(userId); err != nil {
		return 0, err
	}

	var id uint64
	query := "INSERT INTO ssh_keys (user_id, name, public_key) VALUES ($1, $2, $3) RETURNING id"
	err := updateHandler.database.QueryRow(query, userId, name, sanitizedPublicKey).Scan(&id)
	if err != nil {
		return 0, err
	}

	updateHandler.subscriptionHandler.FireSshKeyAddedEvent(userId, id)
	return id, nil
}

func (updateHandler *UpdateHandler) RemoveKey(userId, keyId uint64) error {
	if err := updateHandler.verifier.verifyUserExists(userId); err != nil {
		return err
	}

	query := "DELETE FROM ssh_keys WHERE user_id=$1 AND id=$2"
	result, err := updateHandler.database.Exec(query, userId, keyId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := fmt.Sprintf("Unable to find key for user with id: %d ", keyId)
		return resources.NoSuchKeyError{errorText}
	}

	updateHandler.subscriptionHandler.FireSshKeyRemovedEvent(userId, keyId)
	return nil
}
