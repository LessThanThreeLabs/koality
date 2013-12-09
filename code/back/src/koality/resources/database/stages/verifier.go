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

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyName(name string) error {
	if len(name) < stageMinNameLength {
		return fmt.Errorf("Name must be at least %d characters long", stageMinNameLength)
	} else if len(name) > stageMaxNameLength {
		return fmt.Errorf("Name cannot exceed %d characters long", stageMaxNameLength)
	}
	return nil
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

func (verifier *Verifier) verifyStageDoesNotExistWithSectionAndName(verificationId, sectionNumber uint64, name string) error {
	query := "SELECT id FROM stages WHERE verification_id=$1 AND section_number=$2 AND name=$3"
	err := verifier.database.QueryRow(query, verificationId, sectionNumber, name).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := fmt.Sprintf("Stage already exists with section number %d and name %s", sectionNumber, name)
		return resources.StageAlreadyExistsError{errorText}
	}
	return nil
}
