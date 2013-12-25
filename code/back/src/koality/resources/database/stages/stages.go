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

	createHandler, err := NewCreateHandler(database, verifier)
	if err != nil {
		return nil, err
	}

	readHandler, err := NewReadHandler(database, verifier)
	if err != nil {
		return nil, err
	}

	updateHandler, err := NewUpdateHandler(database, verifier)
	if err != nil {
		return nil, err
	}

	return &resources.StagesHandler{createHandler, readHandler, updateHandler}, nil
}
