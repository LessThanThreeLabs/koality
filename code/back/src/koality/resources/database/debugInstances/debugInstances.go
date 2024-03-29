package debugInstances

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB, buildsHandler *resources.BuildsHandler) (*resources.DebugInstancesHandler, error) {
	internalSubscriptionHandler, err := NewInternalSubscriptionHandler()
	if err != nil {
		return nil, err
	}

	readHandler, err := NewReadHandler(database, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	createHandler, err := NewCreateHandler(database, readHandler, buildsHandler, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	return &resources.DebugInstancesHandler{createHandler, readHandler, internalSubscriptionHandler}, nil
}
