package users

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.UsersHandler, error) {
	createHandler, err := NewCreateHandler(database)
	if err != nil {
		return nil, err
	}

	readHandler, err := NewReadHandler(database)
	if err != nil {
		return nil, err
	}

	updateHandler, err := NewUpdateHandler(database)
	if err != nil {
		return nil, err
	}

	deleteHandler, err := NewDeleteHandler(database)
	if err != nil {
		return nil, err
	}

	return &resources.UsersHandler{createHandler, readHandler, updateHandler, deleteHandler}, nil
}
