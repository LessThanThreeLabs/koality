package stages

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type ReadHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewReadHandler(database *sql.DB, verifier *Verifier) (resources.StagesReadHandler, error) {
	return &ReadHandler{database, verifier}, nil
}

func (readHandler *ReadHandler) Get(stageId uint64) (*resources.Stage, error) {
	stage := new(resources.Stage)
	query := "SELECT id, verification_id, name, flavor, order_number FROM stages WHERE id=$1"
	row := readHandler.database.QueryRow(query, stageId)
	err := row.Scan(&stage.Id, &stage.VerificationId, &stage.Name, &stage.Flavor, &stage.OrderNumber)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage with id: %d", stageId)
		return nil, resources.NoSuchStageError{errorText}
	} else if err != nil {
		return nil, err
	}

	stage.Runs, err = readHandler.GetAllRuns(stageId)
	if err != nil {
		return nil, err
	}

	return stage, nil
}

func (readHandler *ReadHandler) GetAll(verificationId uint64) ([]resources.Stage, error) {
	query := "SELECT id, verification_id, name, flavor, order_number FROM stages WHERE verification_id=$1"
	rows, err := readHandler.database.Query(query, verificationId)
	if err != nil {
		return nil, err
	}

	stages := make([]resources.Stage, 0, 2)
	for rows.Next() {
		stage := resources.Stage{}
		err := rows.Scan(&stage.Id, &stage.VerificationId, &stage.Name, &stage.Flavor, &stage.OrderNumber)
		if err != nil {
			return nil, err
		}
		stages = append(stages, stage)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stages, nil
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

func (readHandler *ReadHandler) GetAllRuns(stageId uint64) ([]resources.StageRun, error) {
	query := "SELECT id, return_code, created, started, ended FROM stage_runs WHERE stage_id=$1"
	rows, err := readHandler.database.Query(query, stageId)
	if err != nil {
		return nil, err
	}

	stageRuns := make([]resources.StageRun, 0, 2)
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
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	lowerBound := offset
	upperBound := offset + results
	query := "SELECT number, text FROM console_texts WHERE run_id=$1 AND number >= $2 AND number < $3 ORDER BY id ASC"
	return readHandler.getConsoleTextLines(query, stageRunId, lowerBound, upperBound)
}

func (readHandler *ReadHandler) GetConsoleTextTail(stageRunId uint64, offset, results int) (map[uint64]string, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

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
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	query := "SELECT number, text FROM console_texts WHERE run_id=$1 ORDER BY id ASC"
	return readHandler.getConsoleTextLines(query, stageRunId)
}

func (readHandler *ReadHandler) scanXunitResult(scannable *sql.Rows) (*resources.XunitResult, error) {
	xunitResult := new(resources.XunitResult)

	var sysout, syserr, failureText, errorText sql.NullString
	err := scannable.Scan(&xunitResult.Name, &xunitResult.Path, &sysout, &syserr, &failureText, &errorText, &xunitResult.Started, &xunitResult.Seconds)
	if err != nil {
		return nil, err
	}

	if sysout.Valid {
		xunitResult.Sysout = sysout.String
	}
	if syserr.Valid {
		xunitResult.Syserr = syserr.String
	}
	if failureText.Valid {
		xunitResult.FailureText = failureText.String
	}
	if errorText.Valid {
		xunitResult.ErrorText = errorText.String
	}

	return xunitResult, nil
}

func (readHandler *ReadHandler) GetXunitResults(stageRunId uint64) ([]resources.XunitResult, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	query := "SELECT name, path, sysout, syserr, failure_text, error_text, started, seconds FROM xunit_results WHERE run_id=$1"
	rows, err := readHandler.database.Query(query, stageRunId)
	if err != nil {
		return nil, err
	}

	xunitResuts := make([]resources.XunitResult, 0, 100)
	for rows.Next() {
		xunitResult, err := readHandler.scanXunitResult(rows)
		if err != nil {
			return nil, err
		}
		xunitResuts = append(xunitResuts, *xunitResult)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return xunitResuts, nil
}

func (readHandler *ReadHandler) GetExports(stageRunId uint64) ([]resources.Export, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	query := "SELECT path, uri FROM exports WHERE run_id=$1"
	rows, err := readHandler.database.Query(query, stageRunId)
	if err != nil {
		return nil, err
	}

	exports := make([]resources.Export, 0, 1)
	for rows.Next() {
		export := resources.Export{}
		err := rows.Scan(&export.Path, &export.Uri)
		if err != nil {
			return nil, err
		}
		exports = append(exports, export)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return exports, nil
}
