package settings

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type DeleteHandler struct {
	database            *sql.DB
	subscriptionHandler resources.InternalSettingsSubscriptionHandler
}

func NewDeleteHandler(database *sql.DB, subscriptionHandler resources.InternalSettingsSubscriptionHandler) (resources.SettingsDeleteHandler, error) {
	return &DeleteHandler{database, subscriptionHandler}, nil
}

func (updateHandler *DeleteHandler) removeSetting(locator SettingLocator) error {
	query := "DELETE FROM settings WHERE key=$1"
	result, err := updateHandler.database.Exec(query, locator.String())
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := fmt.Sprintf("Unable to find setting with locator %v", locator)
		return resources.NoSuchSettingError{errorText}
	}
	return nil
}

func (deleteHandler *DeleteHandler) ClearS3ExporterSettings() error {
	err := deleteHandler.removeSetting(s3ExporterSettingsLocator)
	deleteHandler.subscriptionHandler.FireS3ExporterSettingsClearedEvent()
	return err
}

func (deleteHandler *DeleteHandler) ClearGitHubEnterpriseSettings() error {
	err := deleteHandler.removeSetting(gitHubEnterpriseSettingsLocator)
	deleteHandler.subscriptionHandler.FireGitHubEnterpriseSettingsClearedEvent()
	return err
}
