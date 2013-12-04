package users

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"koality/resources"
)

type UpdateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewUpdateHandler(database *sql.DB) (resources.UsersUpdateHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}
	return &UpdateHandler{database, verifier}, nil
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
	return updateHandler.updateUser(query, firstName, lastName, userId)
}

func (updateHandler *UpdateHandler) SetPassword(userId uint64, passwordHash, passwordSalt []byte) error {
	passwordHashBase64 := base64.StdEncoding.EncodeToString(passwordHash)
	passwordSaltBase64 := base64.StdEncoding.EncodeToString(passwordSalt)

	query := "UPDATE users SET password_hash=$1, password_salt=$2 WHERE id=$3"
	return updateHandler.updateUser(query, passwordHashBase64, passwordSaltBase64, userId)
}

func (updateHandler *UpdateHandler) SetGitHubOauth(userId uint64, gitHubOauth string) error {
	query := "UPDATE users SET github_oauth=$1 WHERE id=$2"
	return updateHandler.updateUser(query, gitHubOauth, userId)
}

func (updateHandler *UpdateHandler) SetAdmin(userId uint64, admin bool) error {
	query := "UPDATE users SET is_admin=$1 WHERE id=$2"
	return updateHandler.updateUser(query, admin, userId)
}

func (updateHandler *UpdateHandler) AddKey(userId uint64, alias, publicKey string) (uint64, error) {
	if err := updateHandler.verifier.verifyKeyAlias(userId, alias); err != nil {
		return 0, err
	} else if err := updateHandler.verifier.verifyPublicKey(publicKey); err != nil {
		return 0, err
	}

	id := uint64(0)
	query := "INSERT INTO ssh_keys (user_id, alias, public_key) VALUES ($1, $2, $3) RETURNING id"
	err := updateHandler.database.QueryRow(query, userId, alias, publicKey).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (updateHandler *UpdateHandler) RemoveKey(userId, keyId uint64) error {
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
	return nil
}
