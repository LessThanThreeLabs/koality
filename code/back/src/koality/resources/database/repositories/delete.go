package repositories

import (
	"database/sql"
	"errors"
	"koality/resources"
)

type DeleteHandler struct {
	database *sql.DB
}

func NewDeleteHandler(database *sql.DB) (resources.RepositoriesDeleteHandler, error) {
	return &DeleteHandler{database}, nil
}

func (deleteHandler *DeleteHandler) Delete(repositoryId uint64) error {
	query := "UPDATE repositories SET deleted = id WHERE id=$1 AND id != deleted"
	result, err := deleteHandler.database.Exec(query, repositoryId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchRepositoryError{errors.New("Unable to find repository")}
	}
	return nil
}