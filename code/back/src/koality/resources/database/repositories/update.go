package repositories

import (
	"database/sql"
	"koality/resources"
	"strings"
)

type UpdateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalRepositoriesSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalRepositoriesSubscriptionHandler) (resources.RepositoriesUpdateHandler, error) {
	return &UpdateHandler{database, verifier, subscriptionHandler}, nil
}

func (updateHandler *UpdateHandler) SetStatus(repositoryId uint64, status string) error {
	if err := updateHandler.verifier.verifyStatus(status); err != nil {
		return err
	}

	query := "UPDATE repositories SET status=$1 WHERE id=$2"
	result, err := updateHandler.database.Exec(query, status, repositoryId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchRepositoryError{"Unable to find repository"}
	}

	updateHandler.subscriptionHandler.FireStatusUpdatedEvent(repositoryId, status)
	return nil
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
		return resources.NoSuchRepositoryHookError{"Unable to find repository hook"}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetGitHubHook(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string) error {
	if err := updateHandler.verifier.verifyHookTypes(hookTypes); err != nil {
		return err
	}

	hookTypesString := strings.Join(hookTypes, ",")
	query := "UPDATE repository_github_metadatas SET hook_id=$1, hook_secret=$2, hook_types=$3 WHERE repository_id=$4"
	err := updateHandler.updateRepositoryHook(query, hookId, hookSecret, hookTypesString, repositoryId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireGitHubHookUpdatedEvent(repositoryId, hookId, hookSecret, hookTypes)
	return nil
}

func (updateHandler *UpdateHandler) ClearGitHubHook(repositoryId uint64) error {
	query := "UPDATE repository_github_metadatas SET hook_id=DEFAULT, hook_secret=DEFAULT, hook_types=DEFAULT WHERE repository_id=$1"
	err := updateHandler.updateRepositoryHook(query, repositoryId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireGitHubHookUpdatedEvent(repositoryId, 0, "", nil)
	return nil
}
