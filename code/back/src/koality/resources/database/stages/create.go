package stages

import (
	"database/sql"
	"koality/resources"
)

const (
	initialVerificationStatus = "received"
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	readHandler         resources.StagesReadHandler
	subscriptionHandler resources.InternalStagesSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.StagesReadHandler,
	subscriptionHandler resources.InternalStagesSubscriptionHandler) (resources.StagesCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) Create(verificationId, sectionNumber uint64, name string, orderNumber uint64) (*resources.Stage, error) {
	err := createHandler.getStageParamsError(verificationId, sectionNumber, name, orderNumber)
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO stages (verification_id, section_number, name, order_number)" +
		" VALUES ($1, $2, $3, $4) RETURNING id"
	err = createHandler.database.QueryRow(query, verificationId, sectionNumber, name, orderNumber).Scan(&id)
	if err != nil {
		return nil, err
	}

	stage, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(stage)
	return stage, nil
}

func (createHandler *CreateHandler) CreateRun(stageId uint64) (*resources.StageRun, error) {
	err := createHandler.getStageRunParamsError(stageId)
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO stage_runs (stage_id) VALUES ($1) RETURNING id"
	err = createHandler.database.QueryRow(query, stageId).Scan(&id)
	if err != nil {
		return nil, err
	}

	stageRun, err := createHandler.readHandler.GetRun(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireRunCreatedEvent(stageRun)
	return stageRun, nil
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
