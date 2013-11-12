package users

import (
	"database/sql"
	"errors"
	"koality/resources"
)

type UpdateHandler struct {
	database *sql.DB
}

func NewUpdateHandler(database *sql.DB) (resources.UsersUpdateHandler, error) {
	return &UpdateHandler{database}, nil
}

func (updateHandler *UpdateHandler) SetName(userId int, firstName, lastName string) error {
	query := "UPDATE users SET first_name=$1, last_name=$2 WHERE id=$3"
	result, err := updateHandler.database.Exec(query, firstName, lastName, userId)
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
