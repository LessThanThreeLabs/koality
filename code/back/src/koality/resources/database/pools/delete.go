package pools

import (
	"database/sql"
	"fmt"
	"koality/resources"
)

type DeleteHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalPoolsSubscriptionHandler
}

func NewDeleteHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalPoolsSubscriptionHandler) (resources.PoolsDeleteHandler, error) {
	return &DeleteHandler{database, verifier, subscriptionHandler}, nil
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

	deleteHandler.subscriptionHandler.FireEc2DeletedEvent(poolId)
	return nil
}
