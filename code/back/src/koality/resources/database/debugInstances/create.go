package debugInstances

import (
	"database/sql"
	"koality/resources"
	"time"
)

type CreateHandler struct {
	database            *sql.DB
	readHandler         resources.DebugInstancesReadHandler
	buildsHandler       *resources.VerificationsHandler
	subscriptionHandler resources.InternalDebugInstancesSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, readHandler resources.DebugInstancesReadHandler, buildsHandler *resources.VerificationsHandler,
	subscriptionHandler resources.InternalDebugInstancesSubscriptionHandler) (resources.DebugInstancesCreateHandler, error) {
	return &CreateHandler{database, readHandler, buildsHandler, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) Create(poolId uint64, instanceId string, expiration *time.Time,
	repositoryInformation *resources.CoreVerificationInformation) (*resources.DebugInstance, error) {
	id := uint64(0)
	query := "INSERT INTO debug_instances (pool_id, instance_id, expires)" +
		" VALUES ($1, $2, $3) RETURNING id"
	err := createHandler.database.QueryRow(query, poolId, instanceId, expiration).Scan(&id)
	if err != nil {
		return nil, err
	}

	_, err = createHandler.buildsHandler.Create.CreateForDebugInstance(
		repositoryInformation.RepositoryId, id, repositoryInformation.HeadSha, repositoryInformation.BaseSha,
		repositoryInformation.HeadMessage, repositoryInformation.HeadUsername, repositoryInformation.HeadEmail,
		repositoryInformation.EmailToNotify)
	if err != nil {
		// TODO(dhuang) on these errors should there be some cleanup?
		return nil, err
	}

	debugInstance, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(debugInstance)
	return debugInstance, nil
}
