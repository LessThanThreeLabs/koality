package snapshots

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"time"
)

var (
	allowedStatuses []string = []string{"running", "succeeded", "failed", "cancelled"}
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyStatus(status string) error {
	for _, allowedBuildStatus := range allowedStatuses {
		if status == allowedBuildStatus {
			return nil
		}
	}
	return resources.InvalidSnapshotStatusError{"Unexpected snapshot status: " + status}
}

func (verifier *Verifier) verifyStartTime(created, started time.Time) error {
	if started.Before(created) {
		return errors.New("Start time cannot be before create time")
	}
	return nil
}

func (verifier *Verifier) verifyEndTime(created, started, ended time.Time) error {
	if started.Before(created) {
		return errors.New("Start time cannot be before create time")
	} else if ended.Before(started) {
		return errors.New("End time cannot be before start time")
	}
	return nil
}

func (verifier *Verifier) verifyPoolExists(poolId uint64) error {
	//TODO(akostov) pools in general? Trying to have snapshots be general over here...
	query := "SELECT id FROM ec2_pools WHERE id=$1"
	err := verifier.database.QueryRow(query, poolId).Scan(new(uint64))
	if err == sql.ErrNoRows {
		return resources.NoSuchPoolError{fmt.Sprintf("Pool with id %d does not exist.", poolId)}
	} else if err != nil {
		return err
	}
	return nil
}
