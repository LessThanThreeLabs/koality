package users

import (
	"database/sql"
	"errors"
	"koality/resources"
)

type DeleteHandler struct {
	database *sql.DB
}

func NewDeleteHandler(database *sql.DB) (resources.UsersDeleteHandler, error) {
	return &DeleteHandler{database}, nil
}

func (deleteHandler *DeleteHandler) Delete(userId int64) error {
	query := "DELETE FROM users WHERE id=$1"
	result, err := deleteHandler.database.Exec(query, userId)
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
