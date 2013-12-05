package users

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type DeleteHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewDeleteHandler(database *sql.DB, verifier *Verifier) (resources.UsersDeleteHandler, error) {
	return &DeleteHandler{database, verifier}, nil
}

func (deleteHandler *DeleteHandler) Delete(userId uint64) error {
	query := "UPDATE users SET deleted = id WHERE id=$1 AND id != deleted"
	result, err := deleteHandler.database.Exec(query, userId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := fmt.Sprintf("Unable to find user with id: %d ", userId)
		return resources.NoSuchUserError{errorText}
	}
	return nil
}
