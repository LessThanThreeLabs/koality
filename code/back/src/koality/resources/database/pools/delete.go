package pools

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type DeleteHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewDeleteHandler(database *sql.DB, verifier *Verifier) (resources.PoolsDeleteHandler, error) {
	return &DeleteHandler{database, verifier}, nil
}

func (deleteHandler *DeleteHandler) DeleteEc2Pool(poolId uint64) error {
	query := "UPDATE ec2_pools SET deleted = id WHERE id=$1 AND id != deleted"
	result, err := deleteHandler.database.Exec(query, poolId)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	} else if count != 1 {
		errorText := fmt.Sprintf("Unable to find pool with id: %d ", poolId)
		return resources.NoSuchPoolError{errorText}
	}
	return nil
}
