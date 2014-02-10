package repositories

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.RepositoriesHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}

	internalSubscriptionHandler, err := NewInternalSubscriptionHandler()
	if err != nil {
		return nil, err
	}

	readHandler, err := NewReadHandler(database, verifier, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	createHandler, err := NewCreateHandler(database, verifier, readHandler, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	updateHandler, err := NewUpdateHandler(database, verifier, readHandler, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	deleteHandler, err := NewDeleteHandler(database, verifier, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	return &resources.RepositoriesHandler{createHandler, readHandler, updateHandler, deleteHandler, internalSubscriptionHandler}, nil
}
