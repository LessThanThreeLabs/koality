package repositories

import (
	"database/sql"
	"errors"
	"koality/resources"
)

type UpdateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewUpdateHandler(database *sql.DB) (resources.RepositoriesUpdateHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}
	return &UpdateHandler{database, verifier}, nil
}

func (updateHandler *UpdateHandler) updateRepositoryHook(query string, params ...interface{}) error {
	result, err := updateHandler.database.Exec(query, params...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchRepositoryHookError{errors.New("Unable to find repository hook")}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetGitHubHook(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string) error {
	// query := "UPDATE repository_github_metadatas SET hook_id=$1, hook_secret=$2, hook_types=$3 WHERE repository_id=$4"
	// return updateHandler.updateRepositoryHook(query, hookId, hookSecret, hookTypes, repositoryId)
	query := "UPDATE repository_github_metadatas SET hook_id=$1, hook_secret=$2 WHERE repository_id=$3"
	return updateHandler.updateRepositoryHook(query, hookId, hookSecret, repositoryId)
}

func (updateHandler *UpdateHandler) ClearGitHubHook(repositoryId uint64) error {
	query := "UPDATE repository_github_metadatas SET hook_id=DEFAULT, hook_secret=DEFAULT, hook_types=DEFAULT WHERE repository_id=$1"
	return updateHandler.updateRepositoryHook(query, repositoryId)
}
