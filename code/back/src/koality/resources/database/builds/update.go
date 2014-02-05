package builds

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
	subscriptionHandler resources.InternalBuildsSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalBuildsSubscriptionHandler) (resources.BuildsUpdateHandler, error) {
	return &UpdateHandler{database, verifier, subscriptionHandler}, nil
}

func (updateHandler *UpdateHandler) updateBuild(query string, params ...interface{}) error {
	result, err := updateHandler.database.Exec(query, params...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchBuildError{"Unable to find build"}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetStatus(buildId uint64, status string) error {
	if err := updateHandler.verifier.verifyStatus(status); err != nil {
		return err
	}

	query := "UPDATE builds SET status=$1 WHERE id=$2"
	err := updateHandler.updateBuild(query, status, buildId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireStatusUpdatedEvent(buildId, status)
	return nil
}

func (updateHandler *UpdateHandler) SetMergeStatus(buildId uint64, mergeStatus string) error {
	if err := updateHandler.verifier.verifyMergeStatus(mergeStatus); err != nil {
		return err
	}

	query := "UPDATE builds SET merge_status=$1 WHERE id=$2"
	err := updateHandler.updateBuild(query, mergeStatus, buildId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireMergeStatusUpdatedEvent(buildId, mergeStatus)
	return nil
}

func (updateHandler *UpdateHandler) getTimes(buildId uint64) (createTime, startTime, endTime *time.Time, err error) {
	query := "SELECT created, started, ended FROM builds WHERE id=$1"
	err = updateHandler.database.QueryRow(query, buildId).Scan(&createTime, &startTime, &endTime)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find build with id: %d", buildId)
		err = resources.NoSuchBuildError{errorText}
	}
	return
}

func (updateHandler *UpdateHandler) SetStartTime(buildId uint64, startTime time.Time) error {
	createTime, _, _, err := updateHandler.getTimes(buildId)
	if err != nil {
		return err
	} else if createTime == nil {
		return errors.New("Cannot set start time when create time is null")
	}

	if err := updateHandler.verifier.verifyStartTime(*createTime, startTime); err != nil {
		return err
	}

	query := "UPDATE builds SET started=$1 WHERE id=$2"
	err = updateHandler.updateBuild(query, startTime, buildId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireStartTimeUpdatedEvent(buildId, startTime)
	return nil
}

func (updateHandler *UpdateHandler) SetEndTime(buildId uint64, endTime time.Time) error {
	createTime, startTime, _, err := updateHandler.getTimes(buildId)
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

	query := "UPDATE builds SET ended=$1 WHERE id=$2"
	err = updateHandler.updateBuild(query, endTime, buildId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireEndTimeUpdatedEvent(buildId, endTime)
	return nil
}
