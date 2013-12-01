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

func NewCreateHandler(database *sql.DB) (resources.StagesCreateHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}
	return &CreateHandler{database, verifier}, nil
}

func (createHandler *CreateHandler) Create(verificationId uint64, name, flavor string, orderNumber uint64) (uint64, error) {
	err := createHandler.getStageParamsError(verificationId, name, flavor, orderNumber)
	if err != nil {
		return 0, err
	}

	id := uint64(0)
	query := "INSERT INTO stages (verification_id, name, flavor, order_number)" +
		" VALUES ($1, $2, $3, $4) RETURNING id"
	err = createHandler.database.QueryRow(query, verificationId, name, flavor, orderNumber).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (createHandler *CreateHandler) CreateRun(stageId uint64) (uint64, error) {
	err := createHandler.getStageRunParamsError(stageId)
	if err != nil {
		return 0, err
	}

	id := uint64(0)
	query := "INSERT INTO stage_runs (stageId) VALUES ($1) RETURNING id"
	err = createHandler.database.QueryRow(query, stageId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (createHandler *CreateHandler) getStageParamsError(verificationId uint64, name, flavor string, orderNumber uint64) error {
	if err := createHandler.verifier.verifyName(name); err != nil {
		return err
	} else if err := createHandler.verifier.verifyFlavor(flavor); err != nil {
		return err
	} else if err := createHandler.verifier.verifyVerificationExists(verificationId); err != nil {
		return err
	} else if err := createHandler.verifier.verifyStageDoesNotExistWithNameAndFlavor(verificationId, name, flavor); err != nil {
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
