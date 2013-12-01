package stages

import (
	"database/sql"
	"koality/resources"
)

func New(database *sql.DB) (*resources.StagesHandler, error) {
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

	return &resources.StagesHandler{createHandler}, nil
}
