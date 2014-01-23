package snapshots

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

const (
	initialSnapshotStatus = "declared"
	defaultImageId        = ""
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

func (createHandler *CreateHandler) Create(poolId uint64, imageType string, repositoryInformation []*resources.SnapshotRepositoryInformation) (*resources.Snapshot, error) {
	fmt.Println("TODO, need to use ", repositoryInformation)
	err := createHandler.getSnapshotParamsError(poolId, imageType)
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO snapshots (pool_id, image_id, image_type, status)" +
		" VALUES ($1, $2, $3, $4) RETURNING id"
	err = createHandler.database.QueryRow(query, poolId, defaultImageId, imageType, initialSnapshotStatus).Scan(&id)
	if err != nil {
		return nil, err
	}

	snapshot, err := createHandler.readHandler.Get(id)
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
