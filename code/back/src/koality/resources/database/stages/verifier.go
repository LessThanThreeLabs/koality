package stages

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"time"
)

const (
	stageMinNameLength = 1
	stageMaxNameLength = 1024
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
	if len(name) < stageMinNameLength {
		errorText := fmt.Sprintf("Name must be at least %d characters long", stageMinNameLength)
		return errors.New(errorText)
	} else if len(name) > stageMaxNameLength {
		errorText := fmt.Sprintf("Name cannot exceed %d characters long", stageMaxNameLength)
		return errors.New(errorText)
	}
	return nil
}

func (verifier *Verifier) verifyFlavor(flavor string) error {
	for _, allowedFlavor := range allowedFlavors {
		if flavor == allowedFlavor {
			return nil
		}
	}
	errorText := "Unexpected stage flavor: " + flavor
	return resources.InvalidStageFlavorError{errorText}
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

func (verifier *Verifier) verifyVerificationExists(verificationId uint64) error {
	query := "SELECT id FROM verifications WHERE id=$1"
	err := verifier.database.QueryRow(query, verificationId).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find verification with id: %d", verificationId)
		return resources.NoSuchVerificationError{errorText}
	}
	return nil
}

func (verifier *Verifier) verifyStageExists(stageId uint64) error {
	query := "SELECT id FROM stages WHERE id=$1"
	err := verifier.database.QueryRow(query, stageId).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage with id: %d", stageId)
		return resources.NoSuchStageError{errorText}
	}
	return nil
}

func (verifier *Verifier) verifyStageRunExists(stageRunId uint64) error {
	query := "SELECT id FROM stage_runs WHERE id=$1"
	err := verifier.database.QueryRow(query, stageRunId).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find stage run with id: %d", stageRunId)
		return resources.NoSuchStageRunError{errorText}
	}
	return nil
}

func (verifier *Verifier) verifyStageDoesNotExistWithNameAndFlavor(verificationId uint64, name, flavor string) error {
	query := "SELECT id FROM stages WHERE verification_id=$1 AND name=$2 AND flavor=$3"
	err := verifier.database.QueryRow(query, verificationId, name, flavor).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := fmt.Sprintf("Stage already exists with name %s and flavor %s", name, flavor)
		return resources.StageAlreadyExistsError{errorText}
	}
	return nil
}
