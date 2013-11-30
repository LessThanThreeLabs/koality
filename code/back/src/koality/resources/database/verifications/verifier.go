package verifications

import (
	"database/sql"
	"errors"
	"koality/resources"
	"regexp"
	"time"
)

const (
	headShaRegex       = "^[a-fA-F0-9]+$"
	baseShaRegex       = "^[a-fA-F0-9]+$"
	emailToNotifyRegex = "^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$"
)

var (
	allowedStatuses      []string = []string{"received", "queued", "running", "passed", "failed", "cancelled"}
	allowedMergeStatuses []string = []string{"running", "passed", "failed"}
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyShas(headSha, baseSha string) error {
	if len(headSha) != 40 {
		return errors.New("Head SHA must be 40 characters long")
	} else if ok, err := regexp.MatchString(headShaRegex, headSha); !ok || err != nil {
		return errors.New("Head SHA must match regex: " + headSha)
	} else if len(baseSha) != 40 {
		return errors.New("Base SHA must be 40 characters long")
	} else if ok, err := regexp.MatchString(baseShaRegex, baseSha); !ok || err != nil {
		return errors.New("Base SHA must match regex: " + baseSha)
	} else if verifier.doesShaPairExist(headSha, baseSha) {
		return resources.ChangesetAlreadyExistsError{errors.New("Changeset already exists with head sha and base sha")}
	}
	return nil
}

func (verifier *Verifier) verifyHeadMessage(message string) error {
	return nil
}

func (verifier *Verifier) verifyHeadUsername(username string) error {
	if len(username) > 128 {
		return errors.New("Head username must be less than 128 characters long")
	}
	return nil
}

func (verifier *Verifier) verifyHeadEmail(email string) error {
	if len(email) > 256 {
		return errors.New("Head email must be less than 256 characters long")
	}
	return nil
}

func (verifier *Verifier) verifyMergeTarget(mergeTarget string) error {
	if len(mergeTarget) > 1024 {
		return errors.New("Merge target must be less than 1024 characters long")
	}
	return nil
}

func (verifier *Verifier) verifyEmailToNotify(emailToNotify string) error {
	if len(emailToNotify) > 256 {
		return errors.New("Email to notify must be less than 256 characters long")
	} else if ok, err := regexp.MatchString(emailToNotifyRegex, emailToNotify); !ok || err != nil {
		return errors.New("Email to notify must match regex: " + emailToNotifyRegex)
	}
	return nil
}

func (verifier *Verifier) verifyStatus(status string) error {
	for _, allowedVerificationStatus := range allowedStatuses {
		if status == allowedVerificationStatus {
			return nil
		}
	}
	return resources.InvalidVerificationStatusError{errors.New("Unexpected verification status: " + status)}
}

func (verifier *Verifier) verifyMergeStatus(mergeStatus string) error {
	for _, allowedMergeStatus := range allowedMergeStatuses {
		if mergeStatus == allowedMergeStatus {
			return nil
		}
	}
	return resources.InvalidVerificationMergeStatusError{errors.New("Unexpected merge status: " + mergeStatus)}
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

func (verifier *Verifier) doesRepositoryExist(repositoryId uint64) bool {
	query := "SELECT id FROM repositories WHERE id=$1"
	err := verifier.database.QueryRow(query, repositoryId).Scan()
	return err != sql.ErrNoRows
}

func (verifier *Verifier) doesChangesetExist(changesetId uint64) bool {
	query := "SELECT id FROM changesets WHERE id=$1"
	err := verifier.database.QueryRow(query, changesetId).Scan()
	return err != sql.ErrNoRows
}

func (verifier *Verifier) doesShaPairExist(headSha, baseSha string) bool {
	query := "SELECT id FROM changesets WHERE head_sha=$1 AND base_sha=$2"
	err := verifier.database.QueryRow(query, headSha, baseSha).Scan()
	return err != sql.ErrNoRows
}
