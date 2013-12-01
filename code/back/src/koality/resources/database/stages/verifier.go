package stages

import (
	"database/sql"
	"errors"
	"koality/resources"
	"time"
)

var (
	allowedFlavors []string = []string{"metadata", "setup", "beforeTest", "test", "afterTest", "deploy"}
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyName(name string) error {
	if len(name) > 1024 {
		return errors.New("Name must be less than 1024 characters long")
	}
	return nil
}

func (verifier *Verifier) verifyFlavor(flavor string) error {
	for _, allowedFlavor := range allowedFlavors {
		if flavor == allowedFlavor {
			return nil
		}
	}
	return resources.InvalidStageFlavorError{errors.New("Unexpected stage flavor: " + flavor)}
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

func (verifier *Verifier) doesVerificationExist(verificationId uint64) bool {
	query := "SELECT id FROM verifications WHERE id=$1"
	err := verifier.database.QueryRow(query, verificationId).Scan()
	return err != sql.ErrNoRows
}

func (verifier *Verifier) doesStageExistWithNameAndFlavor(verificationId uint64, name, flavor string) bool {
	query := "SELECT id FROM stages WHERE verification_id=$1 AND name=$2 AND flavor=$3"
	err := verifier.database.QueryRow(query, verificationId, name, flavor).Scan()
	return err != sql.ErrNoRows
}
