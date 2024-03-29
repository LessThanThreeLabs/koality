package snapshots

import (
	"database/sql"
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
	buildsHandler       *resources.BuildsHandler
	subscriptionHandler resources.InternalSnapshotsSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.SnapshotsReadHandler, buildHandler *resources.BuildsHandler,
	subscriptionHandler resources.InternalSnapshotsSubscriptionHandler) (resources.SnapshotsCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, buildHandler, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) Create(poolId uint64, repositoryInformation []*resources.CoreBuildInformation) (*resources.Snapshot, error) {
	err := createHandler.getSnapshotParamsError(poolId)
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO snapshots (pool_id, image_id, status)" +
		" VALUES ($1, $2, $3) RETURNING id"
	err = createHandler.database.QueryRow(query, poolId, defaultImageId, initialSnapshotStatus).Scan(&id)
	if err != nil {
		return nil, err
	}

	for _, repositoryInformationElement := range repositoryInformation {
		_, err := createHandler.buildsHandler.Create.CreateForSnapshot(repositoryInformationElement.RepositoryId, id,
			repositoryInformationElement.HeadSha, repositoryInformationElement.BaseSha,
			repositoryInformationElement.HeadMessage, repositoryInformationElement.HeadUsername,
			repositoryInformationElement.HeadEmail, repositoryInformationElement.EmailToNotify, repositoryInformationElement.Ref)
		if err != nil {
			return nil, err
		}
	}

	snapshot, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(snapshot)
	return snapshot, nil
}

func (createHandler *CreateHandler) getSnapshotParamsError(poolId uint64) error {
	if err := createHandler.verifier.verifyPoolExists(poolId); err != nil {
		return err
	}
	return nil
}
