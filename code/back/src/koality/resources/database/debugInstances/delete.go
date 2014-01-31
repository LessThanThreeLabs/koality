package debugInstances

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type DeleteHandler struct {
	database             *sql.DB
	verificationsHandler *resources.VerificationsHandler
	subscriptionHandler  resources.InternalDebugInstancesSubscriptionHandler
}

func NewDeleteHandler(database *sql.DB, verificationsHandler *resources.VerificationsHandler,
	subscriptionHandler resources.InternalDebugInstancesSubscriptionHandler) (resources.DebugInstancesDeleteHandler, error) {
	return &DeleteHandler{database, verificationsHandler, subscriptionHandler}, nil
}

func (deleteHandler *DeleteHandler) Delete(debugInstanceId uint64) error {
	query := "UPDATE debug_instances SET deleted = id WHERE id=$1 AND id != deleted"
	result, err := deleteHandler.database.Exec(query, debugInstanceId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := fmt.Sprintf("Unable to find debug instance with id: %d", debugInstanceId)
		return resources.NoSuchDebugInstanceError{errorText}
	}

	deleteHandler.subscriptionHandler.FireDeletedEvent(debugInstanceId)
	return nil
}
