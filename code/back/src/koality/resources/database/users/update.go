package users

import (
	"database/sql"
	"errors"
	"koality/resources"
	"time"
)

type UpdateHandler struct {
	database *sql.DB
}

func NewUpdateHandler(database *sql.DB) (resources.UsersUpdateHandler, error) {
	return &UpdateHandler{database}, nil
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
		return resources.NoSuchUserError(errors.New("Unable to find user"))
	}
	return nil
}

func (updateHandler *UpdateHandler) SetName(userId int64, firstName, lastName string) error {
	query := "UPDATE users SET first_name=$1, last_name=$2 WHERE id=$3"
	return updateHandler.updateUser(query, firstName, lastName, userId)
}

func (updateHandler *UpdateHandler) SetPassword(userId int64, passwordHash, passwordSalt string) error {
	query := "UPDATE users SET password_hash=$1, password_salt=$2 WHERE id=$3"
	return updateHandler.updateUser(query, passwordHash, passwordSalt, userId)
}

func (updateHandler *UpdateHandler) SetGitHubOauth(userId int64, gitHubOauth string) error {
	query := "UPDATE users SET github_oauth=$1 WHERE id=$2"
	return updateHandler.updateUser(query, gitHubOauth, userId)
}

func (updateHandler *UpdateHandler) SetAdmin(userId int64, admin bool) error {
	query := "UPDATE users SET admin=$1 WHERE id=$2"
	return updateHandler.updateUser(query, admin, userId)
}

func (updateHandler *UpdateHandler) AddKey(userId int64, alias, publicKey string) (int64, error) {
	id := int64(0)
	query := "INSERT INTO ssh_keys (user_id, alias, public_key, created) VALUES ($1, $2, $3, $4) RETURNING id"
	err := updateHandler.database.QueryRow(query, userId, alias, publicKey, time.Now()).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (updateHandler *UpdateHandler) RemoveKey(userId, keyId int64) error {
	query := "DELETE FROM ssh_keys WHERE user_id=$1 AND id=$2"
	result, err := updateHandler.database.Exec(query, userId, keyId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchKeyError(errors.New("Unable to find key"))
	}
	return nil
}
