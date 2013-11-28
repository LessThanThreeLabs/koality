package verifications

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.VerificationsHandler, error) {
	createHandler, err := NewCreateHandler(database)
	if err != nil {
		return nil, err
	}

	// readHandler, err := NewReadHandler(database)
	// if err != nil {
	// 	return nil, err
	// }

	// updateHandler, err := NewUpdateHandler(database)
	// if err != nil {
	// 	return nil, err
	// }

	// deleteHandler, err := NewDeleteHandler(database)
	// if err != nil {
	// 	return nil, err
	// }

	// return &resources.VerificationsHandler{createHandler, readHandler, updateHandler, deleteHandler}, nil
	return &resources.VerificationsHandler{createHandler}, nil
}
