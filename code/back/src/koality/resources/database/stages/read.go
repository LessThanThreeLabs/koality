package stages

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type ReadHandler struct {
	database *sql.DB
}

func NewReadHandler(database *sql.DB) (resources.StagesReadHandler, error) {
	return &ReadHandler{database}, nil
}

func (readHandler *ReadHandler) Get(stageId uint64) (*resources.Stage, error) {
	stage := new(resources.Stage)
	query := "SELECT id, verification_id, name, flavor, order_number " +
		"FROM stages WHERE id=$1"
	row := readHandler.database.QueryRow(query, stageId)
	err := row.Scan(&stage.Id, &stage.VerificationId, &stage.Name, &stage.Flavor, &stage.OrderNumber)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage with id: %d", stageId)
		return nil, resources.NoSuchStageError{errorText}
	} else if err != nil {
		return nil, err
	}

	stage.Runs, err = readHandler.GetRuns(stageId)
	if err != nil {
		return nil, err
	}

	return stage, nil
}

func (readHandler *ReadHandler) GetRun(stageRunId uint64) (*resources.StageRun, error) {
	stageRun := new(resources.StageRun)
	query := "SELECT id, return_code, created, started, ended FROM stage_runs WHERE id=$1"
	row := readHandler.database.QueryRow(query, stageRunId)
	err := row.Scan(&stageRun.Id, &stageRun.ReturnCode, &stageRun.Created, &stageRun.Started, &stageRun.Ended)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage run with id: %d", stageRunId)
		return nil, resources.NoSuchStageRunError{errorText}
	} else if err != nil {
		return nil, err
	}
	return stageRun, nil
}

func (readHandler *ReadHandler) GetRuns(stageId uint64) ([]resources.StageRun, error) {
	query := "SELECT id, return_code, created, started, ended FROM stage_runs WHERE stage_id=$1"
	rows, err := readHandler.database.Query(query, stageId)
	if err != nil {
		return nil, err
	}

	stageRuns := make([]resources.StageRun, 0, 1)
	for rows.Next() {
		stageRun := resources.StageRun{}
		err = rows.Scan(&stageRun.Id, &stageRun.ReturnCode, &stageRun.Created, &stageRun.Started, &stageRun.Ended)
		if err != nil {
			return nil, err
		}
		stageRuns = append(stageRuns, stageRun)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stageRuns, nil
}

func (readHandler *ReadHandler) getConsoleTextLines(query string, params ...interface{}) (map[uint64]string, error) {
	rows, err := readHandler.database.Query(query, params...)
	if err != nil {
		return nil, err
	}

	consoleText := make(map[uint64]string)
	for rows.Next() {
		var number uint64
		var text string
		err = rows.Scan(&number, &text)
		if err != nil {
			return nil, err
		}
		consoleText[number] = text
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return consoleText, nil
}

func (readHandler *ReadHandler) GetConsoleTextHead(stageRunId uint64, offset, results int) (map[uint64]string, error) {
	lowerBound := offset
	upperBound := offset + results
	query := "SELECT number, text FROM console_texts WHERE run_id=$1 AND number >= $2 AND number < $3 ORDER BY id ASC"
	return readHandler.getConsoleTextLines(query, stageRunId, lowerBound, upperBound)
}

func (readHandler *ReadHandler) GetConsoleTextTail(stageRunId uint64, offset, results int) (map[uint64]string, error) {
	var maxLineNumber int
	maxLineNumberQuery := "SELECT number FROM console_texts WHERE run_id=$1 ORDER BY number DESC LIMIT 1"
	row := readHandler.database.QueryRow(maxLineNumberQuery, stageRunId)
	err := row.Scan(&maxLineNumber)
	if err == sql.ErrNoRows {
		return map[uint64]string{}, nil
	} else if err != nil {
		return nil, err
	}

	lowerBound := maxLineNumber - offset - results
	upperBound := maxLineNumber - offset
	consoleTextQuery := "SELECT number, text FROM console_texts WHERE run_id=$1 AND number > $2 AND number <= $3 ORDER BY id ASC"
	return readHandler.getConsoleTextLines(consoleTextQuery, stageRunId, lowerBound, upperBound)
}

func (readHandler *ReadHandler) GetAllConsoleText(stageRunId uint64) (map[uint64]string, error) {
	query := "SELECT number, text FROM console_texts WHERE run_id=$1 ORDER BY id ASC"
	return readHandler.getConsoleTextLines(query, stageRunId)
}
