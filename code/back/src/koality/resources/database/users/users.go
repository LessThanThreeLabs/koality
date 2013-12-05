package users

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.UsersHandler, error) {
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

	deleteHandler, err := NewDeleteHandler(database, verifier)
	if err != nil {
		return nil, err
	}

	return &resources.UsersHandler{createHandler, readHandler, updateHandler, deleteHandler}, nil
}
