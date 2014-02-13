package snapshots

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"time"
)

type UpdateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalSnapshotsSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalSnapshotsSubscriptionHandler) (resources.SnapshotsUpdateHandler, error) {
	return &UpdateHandler{database, verifier, subscriptionHandler}, nil
}

func (updateHandler *UpdateHandler) updateSnapshot(query string, params ...interface{}) error {
	result, err := updateHandler.database.Exec(query, params...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchSnapshotError{"Unable to find snapshot"}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetStatus(snapshotId uint64, status string) error {
	if err := updateHandler.verifier.verifyStatus(status); err != nil {
		return err
	}

	query := "UPDATE snapshots SET status=$1 WHERE id=$2"
	err := updateHandler.updateSnapshot(query, status, snapshotId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireStatusUpdatedEvent(snapshotId, status)
	return nil
}

func (updateHandler *UpdateHandler) SetImageId(snapshotId uint64, imageId string) error {
	//TODO(akostov) imageId verifier?

	query := "UPDATE snapshots SET image_id=$1 WHERE id=$2"
	err := updateHandler.updateSnapshot(query, imageId, snapshotId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireImageIdUpdatedEvent(snapshotId, imageId)
	return nil
}

func (updateHandler *UpdateHandler) getTimes(snapshotId uint64) (createTime, startTime, endTime *time.Time, err error) {
	query := "SELECT created, started, ended FROM snapshots WHERE id=$1"
	err = updateHandler.database.QueryRow(query, snapshotId).Scan(&createTime, &startTime, &endTime)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find snapshot with id: %d", snapshotId)
		err = resources.NoSuchSnapshotError{errorText}
	}
	return
}

func (updateHandler *UpdateHandler) SetStartTime(snapshotId uint64, startTime time.Time) error {
	createTime, _, _, err := updateHandler.getTimes(snapshotId)
	if err != nil {
		return err
	} else if createTime == nil {
		return errors.New("Cannot set start time when create time is null")
	}

	if err := updateHandler.verifier.verifyStartTime(*createTime, startTime); err != nil {
		return err
	}

	query := "UPDATE snapshots SET started=$1 WHERE id=$2"
	err = updateHandler.updateSnapshot(query, startTime, snapshotId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireStartTimeUpdatedEvent(snapshotId, startTime)
	return nil
}

func (updateHandler *UpdateHandler) SetEndTime(snapshotId uint64, endTime time.Time) error {
	createTime, startTime, _, err := updateHandler.getTimes(snapshotId)
	if err != nil {
		return err
	} else if createTime == nil {
		return errors.New("Cannot set end time when create time is null")
	} else if startTime == nil {
		return errors.New("Cannot set end time when start time is null")
	}

	if err := updateHandler.verifier.verifyEndTime(*createTime, *startTime, endTime); err != nil {
		return err
	}

	query := "UPDATE snapshots SET ended=$1 WHERE id=$2"
	err = updateHandler.updateSnapshot(query, endTime, snapshotId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireEndTimeUpdatedEvent(snapshotId, endTime)
	return nil
}
