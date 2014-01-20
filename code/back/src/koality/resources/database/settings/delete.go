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

func (updateHandler *DeleteHandler) removeSetting(resource, key string) error {
	query := "DELETE FROM settings WHERE resource=$1 AND key=$2"
	result, err := updateHandler.database.Exec(query, resource, key)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := fmt.Sprintf("Unable to find setting %s-%s", resource, key)
		return resources.NoSuchSettingError{errorText}
	}
	return nil
}

func (deleteHandler *DeleteHandler) ClearS3ExporterSettings() error {
	err := deleteHandler.removeSetting("Exporter", "S3Settings")
	deleteHandler.subscriptionHandler.FireS3ExporterSettingsClearedEvent()
	return err
}
