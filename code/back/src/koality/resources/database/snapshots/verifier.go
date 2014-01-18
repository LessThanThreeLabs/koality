package snapshots

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"time"
)

var (
	allowedImageTypes []string = []string{"ec2"}
	allowedStatuses   []string = []string{"running", "succeeded", "failed", "cancelled"}
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyImageType(imageType string) error {
	for _, allowedImageType := range allowedImageTypes {
		if imageType == allowedImageType {
			return nil
		}
	}
	return fmt.Errorf("Instance type must be one of: %v", allowedImageTypes)
}

func (verifier *Verifier) verifyStatus(status string) error {
	for _, allowedVerificationStatus := range allowedStatuses {
		if status == allowedVerificationStatus {
			return nil
		}
	}
	return resources.InvalidVerificationStatusError{"Unexpected verification status: " + status}
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
	if err != nil {
		return err
	} else if err == sql.ErrNoRows {
		return resources.PoolDoesNotExistError{fmt.Sprintf("Pool with id %d does not exist.", poolId)}
	}
	return nil
}
