package stages

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type ReadHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalStagesSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalStagesSubscriptionHandler) (resources.StagesReadHandler, error) {
	return &ReadHandler{database, verifier, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) Get(stageId uint64) (*resources.Stage, error) {
	stage := new(resources.Stage)
	query := "SELECT id, build_id, section_number, name, order_number FROM stages WHERE id=$1"
	row := readHandler.database.QueryRow(query, stageId)
	err := row.Scan(&stage.Id, &stage.BuildId, &stage.SectionNumber, &stage.Name, &stage.OrderNumber)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage with id: %d", stageId)
		return nil, resources.NoSuchStageError{errorText}
	} else if err != nil {
		return nil, err
	}

	stage.Runs, err = readHandler.GetAllRuns(stage.Id)
	if err != nil {
		return nil, err
	}

	return stage, nil
}

func (readHandler *ReadHandler) GetBySectionNumberAndName(buildId, sectionNumber uint64, name string) (*resources.Stage, error) {
	stage := new(resources.Stage)
	query := "SELECT id, build_id, section_number, name, order_number FROM stages WHERE build_id=$1 AND section_number=$2 AND name=$3"
	row := readHandler.database.QueryRow(query, buildId, sectionNumber, name)
	err := row.Scan(&stage.Id, &stage.BuildId, &stage.SectionNumber, &stage.Name, &stage.OrderNumber)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage with section number %d and name %s", sectionNumber, name)
		return nil, resources.NoSuchStageError{errorText}
	} else if err != nil {
		return nil, err
	}

	stage.Runs, err = readHandler.GetAllRuns(stage.Id)
	if err != nil {
		return nil, err
	}

	return stage, nil
}

func (readHandler *ReadHandler) GetAll(buildId uint64) ([]resources.Stage, error) {
	query := "SELECT id, build_id, section_number, name, order_number FROM stages WHERE build_id=$1"
	rows, err := readHandler.database.Query(query, buildId)
	if err != nil {
		return nil, err
	}

	stages := make([]resources.Stage, 0, 2)
	for rows.Next() {
		stage := resources.Stage{}
		err := rows.Scan(&stage.Id, &stage.BuildId, &stage.SectionNumber, &stage.Name, &stage.OrderNumber)
		if err != nil {
			return nil, err
		}

		stage.Runs, err = readHandler.GetAllRuns(stage.Id)
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
	query := "SELECT id, stage_id, return_code, created, started, ended FROM stage_runs WHERE id=$1"
	row := readHandler.database.QueryRow(query, stageRunId)
	err := row.Scan(&stageRun.Id, &stageRun.StageId, &stageRun.ReturnCode, &stageRun.Created, &stageRun.Started, &stageRun.Ended)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage run with id: %d", stageRunId)
		return nil, resources.NoSuchStageRunError{errorText}
	} else if err != nil {
		return nil, err
	}

	var hasConsoleLines, hasXunitResults, hasExports int
	hasConsoleLinesQuery := "SELECT COUNT(*) FROM (SELECT 1 FROM console_lines WHERE run_id=$1 LIMIT 1) AS _"
	hasXunitResultsQuery := "SELECT COUNT(*) FROM (SELECT 1 FROM xunit_results WHERE run_id=$1 LIMIT 1) AS _"
	hasExportsQuery := "SELECT COUNT(*) FROM (SELECT 1 FROM exports WHERE run_id=$1 LIMIT 1) AS _"

	if err := readHandler.database.QueryRow(hasConsoleLinesQuery, stageRunId).Scan(&hasConsoleLines); err != nil {
		return nil, err
	} else if err := readHandler.database.QueryRow(hasXunitResultsQuery, stageRunId).Scan(&hasXunitResults); err != nil {
		return nil, err
	} else if err := readHandler.database.QueryRow(hasExportsQuery, stageRunId).Scan(&hasExports); err != nil {
		return nil, err
	}

	stageRun.HasConsoleLines = hasConsoleLines == 1
	stageRun.HasXunitResults = hasXunitResults == 1
	stageRun.HasExports = hasExports == 1
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

func (readHandler *ReadHandler) getConsoleLines(query string, params ...interface{}) (map[uint64]string, error) {
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

func (readHandler *ReadHandler) GetConsoleLinesHead(stageRunId uint64, offset, results uint32) (map[uint64]string, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	lowerBound := offset
	upperBound := offset + results
	query := "SELECT number, text FROM console_lines WHERE run_id=$1 AND number >= $2 AND number < $3 ORDER BY id ASC"
	return readHandler.getConsoleLines(query, stageRunId, lowerBound, upperBound)
}

func (readHandler *ReadHandler) GetConsoleLinesTail(stageRunId uint64, offset, results uint32) (map[uint64]string, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	var maxLineNumber uint32
	maxLineNumberQuery := "SELECT number FROM console_lines WHERE run_id=$1 ORDER BY number DESC LIMIT 1"
	row := readHandler.database.QueryRow(maxLineNumberQuery, stageRunId)
	err := row.Scan(&maxLineNumber)
	if err == sql.ErrNoRows {
		return map[uint64]string{}, nil
	} else if err != nil {
		return nil, err
	}

	if maxLineNumber < offset {
		return map[uint64]string{}, nil
	}

	var lowerBound uint32
	if maxLineNumber < offset+results {
		lowerBound = 0
	} else {
		lowerBound = maxLineNumber - offset - results
	}

	upperBound := maxLineNumber - offset

	consoleTextQuery := "SELECT number, text FROM console_lines WHERE run_id=$1 AND number > $2 AND number <= $3 ORDER BY id ASC"
	return readHandler.getConsoleLines(consoleTextQuery, stageRunId, lowerBound, upperBound)
}

func (readHandler *ReadHandler) GetAllConsoleLines(stageRunId uint64) (map[uint64]string, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	query := "SELECT number, text FROM console_lines WHERE run_id=$1 ORDER BY id ASC"
	return readHandler.getConsoleLines(query, stageRunId)
}

func (readHandler *ReadHandler) scanXunitResult(scannable *sql.Rows) (*resources.XunitResult, error) {
	xunitResult := new(resources.XunitResult)

	var sysout, syserr, failureText, errorText sql.NullString
	err := scannable.Scan(&xunitResult.Name, &xunitResult.Path, &sysout, &syserr, &failureText, &errorText, &xunitResult.Seconds)
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

func (readHandler *ReadHandler) GetAllXunitResults(stageRunId uint64) ([]resources.XunitResult, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	query := "SELECT name, path, sysout, syserr, failure_text, error_text, seconds FROM xunit_results WHERE run_id=$1"
	rows, err := readHandler.database.Query(query, stageRunId)
	if err != nil {
		return nil, err
	}

	xunitResults := make([]resources.XunitResult, 0, 100)
	for rows.Next() {
		xunitResult, err := readHandler.scanXunitResult(rows)
		if err != nil {
			return nil, err
		}
		xunitResults = append(xunitResults, *xunitResult)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return xunitResults, nil
}

func (readHandler *ReadHandler) GetAllExports(stageRunId uint64) ([]resources.Export, error) {
	if err := readHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return nil, err
	}

	query := "SELECT bucket, path, key FROM exports WHERE run_id=$1"
	rows, err := readHandler.database.Query(query, stageRunId)
	if err != nil {
		return nil, err
	}

	exports := make([]resources.Export, 0, 1)
	for rows.Next() {
		export := resources.Export{}
		err := rows.Scan(&export.BucketName, &export.Path, &export.Key)
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
