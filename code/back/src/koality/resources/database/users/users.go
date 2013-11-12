package users

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.UsersHandler, error) {
	readHandler, err := NewReadHandler(database)
	if err != nil {
		return nil, err
	}

	updateHandler, err := NewUpdateHandler(database)
	if err != nil {
		return nil, err
	}

	return &resources.UsersHandler{readHandler, updateHandler}, nil
}
