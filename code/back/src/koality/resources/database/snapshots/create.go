package snapshots

import (
	"database/sql"
	"koality/resources"
)

const (
	initialSnapshotStatus = "declared"
	defaultImageId        = ""
	initiallyDeleted      = false
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	readHandler         resources.SnapshotsReadHandler
	subscriptionHandler resources.InternalSnapshotsSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.SnapshotsReadHandler,
	subscriptionHandler resources.InternalSnapshotsSubscriptionHandler) (resources.SnapshotsCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) CreateSnapshot(poolId uint64, imageType string) (*resources.Snapshot, error) {
	err := createHandler.getSnapshotParamsError(poolId, imageType)
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO snapshots (poolId, imageId, imageType, status, deleted)" +
		" VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = createHandler.database.QueryRow(query, poolId, defaultImageId, imageType, initialSnapshotStatus, initiallyDeleted).Scan(&id)
	if err != nil {
		return nil, err
	}

	snapshot, err := createHandler.readHandler.GetSnapshot(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(snapshot)
	return snapshot, nil
}

func (createHandler *CreateHandler) getSnapshotParamsError(poolId uint64, imageType string) error {
	if err := createHandler.verifier.verifyImageType(imageType); err != nil {
		return err
	} else if err := createHandler.verifier.verifyPoolExists(poolId); err != nil {
		return err
	}
	return nil
}
