package stages

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"time"
)

type UpdateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewUpdateHandler(database *sql.DB) (resources.StagesUpdateHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}
	return &UpdateHandler{database, verifier}, nil
}

func (updateHandler *UpdateHandler) updateStageRun(query string, params ...interface{}) error {
	result, err := updateHandler.database.Exec(query, params...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := "Unable to find stage run"
		return resources.NoSuchStageRunError{errors.New(errorText)}
	}
	return nil
}

func (updateHandler *UpdateHandler) SetReturnCode(stageRunId uint64, returnCode int) error {
	query := "UPDATE stage_runs SET return_code=$1 WHERE id=$2"
	return updateHandler.updateStageRun(query, returnCode, stageRunId)
}

func (updateHandler *UpdateHandler) getTimes(stageRunId uint64) (createTime, startTime, endTime *time.Time, err error) {
	query := "SELECT created, started, ended FROM stage_runs WHERE id=$1"
	err = updateHandler.database.QueryRow(query, stageRunId).Scan(&createTime, &startTime, &endTime)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage run with id: %d", stageRunId)
		err = resources.NoSuchVerificationError{errors.New(errorText)}
	}
	return
}

func (updateHandler *UpdateHandler) SetStartTime(stageRunId uint64, startTime time.Time) error {
	createTime, _, _, err := updateHandler.getTimes(stageRunId)
	if err != nil {
		return err
	} else if createTime == nil {
		return errors.New("Cannot set start time when create time is null")
	}

	if err := updateHandler.verifier.verifyStartTime(*createTime, startTime); err != nil {
		return err
	}
	query := "UPDATE stage_runs SET started=$1 WHERE id=$2"
	return updateHandler.updateStageRun(query, startTime, stageRunId)
}

func (updateHandler *UpdateHandler) SetEndTime(stageRunId uint64, endTime time.Time) error {
	createTime, startTime, _, err := updateHandler.getTimes(stageRunId)
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
	query := "UPDATE stage_runs SET ended=$1 WHERE id=$2"
	return updateHandler.updateStageRun(query, endTime, stageRunId)
}

// func (updateHandler *UpdateHandler) AddKey(userId uint64, alias, publicKey string) (uint64, error) {
// 	if err := updateHandler.verifier.verifyKeyAlias(userId, alias); err != nil {
// 		return 0, err
// 	} else if err := updateHandler.verifier.verifyPublicKey(publicKey); err != nil {
// 		return 0, err
// 	}

// 	id := uint64(0)
// 	query := "INSERT INTO ssh_keys (user_id, alias, public_key) VALUES ($1, $2, $3) RETURNING id"
// 	err := updateHandler.database.QueryRow(query, userId, alias, publicKey).Scan(&id)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return id, nil
// }
