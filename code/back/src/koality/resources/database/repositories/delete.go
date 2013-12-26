package repositories

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type DeleteHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalRepositoriesSubscriptionHandler
}

func NewDeleteHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalRepositoriesSubscriptionHandler) (resources.RepositoriesDeleteHandler, error) {
	return &DeleteHandler{database, verifier, subscriptionHandler}, nil
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
		errorText := fmt.Sprintf("Unable to find repository with id: %d ", repositoryId)
		return resources.NoSuchRepositoryError{errorText}
	}

	deleteHandler.subscriptionHandler.FireDeletedEvent(repositoryId)
	return nil
}
