package users

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.UsersHandler, error) {
	readHandler, err := NewReadHandler(database)
	// if err != nil {
	// 	return nil, err
	// }
	var _ = err

	return &resources.UsersHandler{readHandler}, nil
}
