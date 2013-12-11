package stages

import (
	"database/sql"
	"koality/resources"
)

const (
	initialVerificationStatus = "received"
)

type CreateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewCreateHandler(database *sql.DB, verifier *Verifier) (resources.StagesCreateHandler, error) {
	return &CreateHandler{database, verifier}, nil
}

func (createHandler *CreateHandler) Create(verificationId, sectionNumber uint64, name string, orderNumber uint64) (uint64, error) {
	err := createHandler.getStageParamsError(verificationId, sectionNumber, name, orderNumber)
	if err != nil {
		return 0, err
	}

	id := uint64(0)
	query := "INSERT INTO stages (verification_id, section_number, name, order_number)" +
		" VALUES ($1, $2, $3, $4) RETURNING id"
	err = createHandler.database.QueryRow(query, verificationId, sectionNumber, name, orderNumber).Scan(&id)
	return id, err
}

func (createHandler *CreateHandler) CreateRun(stageId uint64) (uint64, error) {
	err := createHandler.getStageRunParamsError(stageId)
	if err != nil {
		return 0, err
	}

	id := uint64(0)
	query := "INSERT INTO stage_runs (stage_id) VALUES ($1) RETURNING id"
	err = createHandler.database.QueryRow(query, stageId).Scan(&id)
	return id, err
}

func (createHandler *CreateHandler) getStageParamsError(verificationId, sectionNumber uint64, name string, orderNumber uint64) error {
	if err := createHandler.verifier.verifyName(name); err != nil {
		return err
	} else if err := createHandler.verifier.verifyVerificationExists(verificationId); err != nil {
		return err
	} else if err := createHandler.verifier.verifyStageDoesNotExistWithSectionAndName(verificationId, sectionNumber, name); err != nil {
		return err
	}
	return nil
}

func (createHandler *CreateHandler) getStageRunParamsError(stageId uint64) error {
	if err := createHandler.verifier.verifyStageExists(stageId); err != nil {
		return err
	}
	return nil
}
