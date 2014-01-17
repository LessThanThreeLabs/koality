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
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalStagesSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalStagesSubscriptionHandler) (resources.StagesUpdateHandler, error) {
	return &UpdateHandler{database, verifier, subscriptionHandler}, nil
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
	err := updateHandler.updateStageRun(query, returnCode, stageRunId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireReturnCodeUpdatedEvent(stageRunId, returnCode)
	return nil
}

func (updateHandler *UpdateHandler) getTimes(stageRunId uint64) (createTime, startTime, endTime *time.Time, err error) {
	query := "SELECT created, started, ended FROM stage_runs WHERE id=$1"
	err = updateHandler.database.QueryRow(query, stageRunId).Scan(&createTime, &startTime, &endTime)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage run with id: %d", stageRunId)
		err = resources.NoSuchStageRunError{errorText}
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
	err = updateHandler.updateStageRun(query, startTime, stageRunId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireStartTimeUpdatedEvent(stageRunId, startTime)
	return nil
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
	err = updateHandler.updateStageRun(query, endTime, stageRunId)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireEndTimeUpdatedEvent(stageRunId, endTime)
	return nil
}

func (updateHandler *UpdateHandler) AddConsoleLines(stageRunId uint64, consoleLines map[uint64]string) error {
	// We don't verify that stageRunId exists for performance reasons

	if len(consoleLines) == 0 {
		return nil
	}

	getValuesString := func() string {
		valuesStringArray := make([]string, len(consoleLines))
		for index := 0; index < len(consoleLines); index++ {
			valuesStringArray[index] = fmt.Sprintf("(%d, $%d, $%d)", stageRunId, index*2+1, index*2+2)
		}
		return strings.Join(valuesStringArray, ", ")
	}

	consoleLinesToArray := func() []interface{} {
		count := 0
		conoleLinesArray := make([]interface{}, len(consoleLines)*2)
		for number, text := range consoleLines {
			conoleLinesArray[count*2] = number
			conoleLinesArray[count*2+1] = text
			count++
		}
		return conoleLinesArray
	}

	query := "INSERT INTO console_lines (run_id, number, text) VALUES " + getValuesString()
	_, err := updateHandler.database.Exec(query, consoleLinesToArray()...)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireConsoleLinesAddedEvent(stageRunId, consoleLines)
	return nil
}

func (updateHandler *UpdateHandler) RemoveAllConsoleLines(stageRunId uint64) error {
	if err := updateHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return err
	}

	query := "DELETE FROM console_lines WHERE run_id=$1"
	_, err := updateHandler.database.Exec(query, stageRunId)
	return err
}

func (updateHandler *UpdateHandler) AddXunitResults(stageRunId uint64, xunitResults []resources.XunitResult) error {
	if err := updateHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return err
	}

	if len(xunitResults) == 0 {
		return nil
	}

	getValuesString := func() string {
		valuesStringArray := make([]string, len(xunitResults))
		for index := 0; index < len(xunitResults); index++ {
			valuesStringArray[index] = fmt.Sprintf("(%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				stageRunId, index*7+1, index*7+2, index*7+3, index*7+4, index*7+5, index*7+6, index*7+7)
		}
		return strings.Join(valuesStringArray, ", ")
	}

	xunitResultsToArray := func() []interface{} {
		xunitResultsArray := make([]interface{}, len(xunitResults)*7)
		for index, xunitResult := range xunitResults {
			xunitResultsArray[index*7] = xunitResult.Name
			xunitResultsArray[index*7+1] = xunitResult.Path
			xunitResultsArray[index*7+2] = xunitResult.Sysout
			xunitResultsArray[index*7+3] = xunitResult.Syserr
			xunitResultsArray[index*7+4] = xunitResult.FailureText
			xunitResultsArray[index*7+5] = xunitResult.ErrorText
			xunitResultsArray[index*7+6] = xunitResult.Seconds
		}
		return xunitResultsArray
	}

	query := "INSERT INTO xunit_results (run_id, name, path, sysout, syserr, failure_text, error_text, seconds) VALUES " + getValuesString()
	_, err := updateHandler.database.Exec(query, xunitResultsToArray()...)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireXunitResultsAddedEvent(stageRunId, xunitResults)
	return nil
}

func (updateHandler *UpdateHandler) RemoveAllXunitResults(stageRunId uint64) error {
	if err := updateHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return err
	}

	query := "DELETE FROM xunit_results WHERE run_id=$1"
	_, err := updateHandler.database.Exec(query, stageRunId)
	return err
}

func (updateHandler *UpdateHandler) AddExports(stageRunId uint64, exports []resources.Export) error {
	if len(exports) == 0 {
		return nil
	}

	if err := updateHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return err
	}

	if len(exports) == 0 {
		return nil
	}

	getValuesString := func() string {
		valuesStringArray := make([]string, len(exports))
		for index := 0; index < len(exports); index++ {
			valuesStringArray[index] = fmt.Sprintf("(%d, $%d, $%d, $%d)", stageRunId, index*3+1, index*3+2, index*3+3)
		}
		return strings.Join(valuesStringArray, ", ")
	}

	exportsToArray := func() []interface{} {
		exportsArray := make([]interface{}, len(exports)*3)
		for index, export := range exports {
			exportsArray[index*3] = export.BucketName
			exportsArray[index*3+1] = export.Path
			exportsArray[index*3+2] = export.Key
		}
		return exportsArray
	}

	query := "INSERT INTO exports (run_id, bucket, path, key) VALUES " + getValuesString()
	_, err := updateHandler.database.Exec(query, exportsToArray()...)
	if err != nil {
		return err
	}

	updateHandler.subscriptionHandler.FireExportsAddedEvent(stageRunId, exports)
	return nil
}
