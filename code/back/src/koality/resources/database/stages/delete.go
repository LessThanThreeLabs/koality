package stages

import (
	"database/sql"
	"koality/resources"
)

type DeleteHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewDeleteHandler(database *sql.DB, verifier *Verifier) (resources.StagesDeleteHandler, error) {
	return &DeleteHandler{database, verifier}, nil
}

func (deleteHandler *DeleteHandler) DeleteAllConsoleLines(stageRunId uint64) error {
	if err := deleteHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return err
	}

	query := "DELETE FROM console_lines WHERE run_id=$1"
	_, err := deleteHandler.database.Exec(query, stageRunId)
	return err
}

func (deleteHandler *DeleteHandler) DeleteAllXunitResults(stageRunId uint64) error {
	if err := deleteHandler.verifier.verifyStageRunExists(stageRunId); err != nil {
		return err
	}

	query := "DELETE FROM xunit_results WHERE run_id=$1"
	_, err := deleteHandler.database.Exec(query, stageRunId)
	return err
}
