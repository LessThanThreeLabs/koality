package snapshots

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type DeleteHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalSnapshotsSubscriptionHandler
}

func NewDeleteHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalSnapshotsSubscriptionHandler) (resources.SnapshotsDeleteHandler, error) {
	return &DeleteHandler{database, verifier, subscriptionHandler}, nil
}

func (deleteHandler *DeleteHandler) DeleteSnapshot(snapshotId uint64) error {
	query := "UPDATE ec2_snapshots SET deleted = id WHERE id=$1 AND id != deleted"
	result, err := deleteHandler.database.Exec(query, snapshotId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := fmt.Sprintf("Unable to find snapshot with id: %d ", snapshotId)
		return resources.NoSuchPoolError{errorText}
	}

	deleteHandler.subscriptionHandler.FireDeletedEvent(snapshotId)
	return nil
}
