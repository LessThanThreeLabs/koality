package stages

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.StagesHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}

	internalSubscriptionHandler, err := NewInternalSubscriptionHandler()
	if err != nil {
		return nil, err
	}

	createHandler, err := NewCreateHandler(database, verifier, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	readHandler, err := NewReadHandler(database, verifier, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	updateHandler, err := NewUpdateHandler(database, verifier, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	return &resources.StagesHandler{createHandler, readHandler, updateHandler, internalSubscriptionHandler}, nil
}
