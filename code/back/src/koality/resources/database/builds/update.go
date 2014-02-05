package verifications

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
	subscriptionHandler resources.InternalVerificationsSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalVerificationsSubscriptionHandler) (resources.VerificationsUpdateHandler, error) {
	return &UpdateHandler{database, verifier, subscriptionHandler}, nil
}

func (updateHandler *UpdateHandler) updateVerification(query string, params ...interface{}) error {
	result, err := updateHandler.database.Exec(query, params...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		return resources.NoSuchVerificationError{"Unable to find verification"}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetStatus(verificationId uint64, status string) error {
	if err := updateHandler.verifier.verifyStatus(status); err != nil {
		return err
	}

	query := "UPDATE verifications SET status=$1 WHERE id=$2"
	err := updateHandler.updateVerification(query, status, verificationId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireStatusUpdatedEvent(verificationId, status)
	return nil
}

func (updateHandler *UpdateHandler) SetMergeStatus(verificationId uint64, mergeStatus string) error {
	if err := updateHandler.verifier.verifyMergeStatus(mergeStatus); err != nil {
		return err
	}

	query := "UPDATE verifications SET merge_status=$1 WHERE id=$2"
	err := updateHandler.updateVerification(query, mergeStatus, verificationId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireMergeStatusUpdatedEvent(verificationId, mergeStatus)
	return nil
}

func (updateHandler *UpdateHandler) getTimes(verificationId uint64) (createTime, startTime, endTime *time.Time, err error) {
	query := "SELECT created, started, ended FROM verifications WHERE id=$1"
	err = updateHandler.database.QueryRow(query, verificationId).Scan(&createTime, &startTime, &endTime)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find verification with id: %d", verificationId)
		err = resources.NoSuchVerificationError{errorText}
	}
	return
}

func (updateHandler *UpdateHandler) SetStartTime(verificationId uint64, startTime time.Time) error {
	createTime, _, _, err := updateHandler.getTimes(verificationId)
	if err != nil {
		return err
	} else if createTime == nil {
		return errors.New("Cannot set start time when create time is null")
	}

	if err := updateHandler.verifier.verifyStartTime(*createTime, startTime); err != nil {
		return err
	}

	query := "UPDATE verifications SET started=$1 WHERE id=$2"
	err = updateHandler.updateVerification(query, startTime, verificationId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireStartTimeUpdatedEvent(verificationId, startTime)
	return nil
}

func (updateHandler *UpdateHandler) SetEndTime(verificationId uint64, endTime time.Time) error {
	createTime, startTime, _, err := updateHandler.getTimes(verificationId)
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

	query := "UPDATE verifications SET ended=$1 WHERE id=$2"
	err = updateHandler.updateVerification(query, endTime, verificationId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireEndTimeUpdatedEvent(verificationId, endTime)
	return nil
}
