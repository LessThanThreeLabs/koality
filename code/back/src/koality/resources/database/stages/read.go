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

func (readHandler *ReadHandler) GetConsoleText(stageRunId uint64, offset, results uint64) (map[uint64]string, error) {
	// query := "SELECT number, line FROM console_texts WHERE run_id=$1 ORDER BY number ASC LIMIT $2 OFFSET $3"
	return nil, nil
}

func (readHandler *ReadHandler) GetConsoleTextTail(stageRunId uint64, offset, results uint64) (map[uint64]string, error) {
	// query := "SELECT number, line FROM console_texts WHERE run_id=$1 ORDER BY number DESC LIMIT $2 OFFSET $3"
	return nil, nil
}

func (readHandler *ReadHandler) GetAllConsoleText(stageRunId uint64) (map[uint64]string, error) {
	// query := "SELECT number, line FROM console_texts WHERE run_id=$1"
	return nil, nil
}
