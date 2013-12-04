package stages

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"strings"
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
		return resources.NoSuchStageRunError{"Unable to find stage run"}
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
		err = resources.NoSuchVerificationError{errorText}
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

func (updateHandler *UpdateHandler) AddConsoleLines(stageRunId uint64, consoleTextLines ...resources.ConsoleTextLine) error {
	// We don't verify that stageRunId exists for performance reasons

	getValuesString := func() string {
		valuesStringArray := make([]string, len(consoleTextLines))
		for index, _ := range consoleTextLines {
			valuesStringArray[index] = fmt.Sprintf("(%d, $%d, $%d)", stageRunId, index*2+1, index*2+2)
		}
		return strings.Join(valuesStringArray, ", ")
	}

	consoleTextLinesToArray := func() []interface{} {
		linesArray := make([]interface{}, len(consoleTextLines)*2)
		for index, consoleTextLine := range consoleTextLines {
			linesArray[index*2] = consoleTextLine.Number
			linesArray[index*2+1] = consoleTextLine.Text
		}
		return linesArray
	}

	query := "INSERT INTO console_texts (run_id, number, text) VALUES " + getValuesString()
	_, err := updateHandler.database.Exec(query, consoleTextLinesToArray()...)
	if err != nil {
		return err
	}
	return nil
}
