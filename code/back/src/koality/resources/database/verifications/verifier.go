package verifications

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"regexp"
	"time"
)

const (
	verificationHeadShaLength          = 40
	verificationBaseShaLength          = 40
	verificationMaxHeadUsernameLength  = 128
	verificationMaxHeadEmailLength     = 256
	verificationMaxMergeTargetLength   = 1024
	verificationMaxEmailToNotifyLength = 128
	headShaRegex                       = "^[a-fA-F0-9]+$"
	baseShaRegex                       = "^[a-fA-F0-9]+$"
	emailToNotifyRegex                 = "^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$"
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

func (verifier *Verifier) verifyHeadSha(headSha string) error {
	if len(headSha) != verificationHeadShaLength {
		return fmt.Errorf("Head SHA must be %d characters long", verificationHeadShaLength)
	} else if ok, err := regexp.MatchString(headShaRegex, headSha); !ok || err != nil {
		return errors.New("Head SHA must match regex: " + headSha)
	}
	return nil
}

func (verifier *Verifier) verifyBaseSha(baseSha string) error {
	if len(baseSha) != verificationBaseShaLength {
		return fmt.Errorf("Base SHA must be %d characters long", verificationBaseShaLength)
	} else if ok, err := regexp.MatchString(baseShaRegex, baseSha); !ok || err != nil {
		return errors.New("Base SHA must match regex: " + baseSha)
	}
	return nil
}

func (verifier *Verifier) verifyHeadMessage(message string) error {
	return nil
}

func (verifier *Verifier) verifyHeadUsername(username string) error {
	if len(username) > verificationMaxHeadUsernameLength {
		return fmt.Errorf("Head username cannot exceed %d characters long", verificationMaxHeadUsernameLength)
	}
	return nil
}

func (verifier *Verifier) verifyHeadEmail(email string) error {
	if len(email) > verificationMaxHeadEmailLength {
		return fmt.Errorf("Head email cannot exceed %d characters long", verificationMaxHeadEmailLength)
	}
	return nil
}

func (verifier *Verifier) verifyMergeTarget(mergeTarget string) error {
	if len(mergeTarget) > verificationMaxMergeTargetLength {
		return fmt.Errorf("Merge target cannot exceed %d characters long", verificationMaxMergeTargetLength)
	}
	return nil
}

func (verifier *Verifier) verifyEmailToNotify(emailToNotify string) error {
	if len(emailToNotify) > verificationMaxEmailToNotifyLength {
		return fmt.Errorf("Email to notify cannot exceed %d characters long", verificationMaxEmailToNotifyLength)
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
	return resources.InvalidVerificationStatusError{"Unexpected verification status: " + status}
}

func (verifier *Verifier) verifyMergeStatus(mergeStatus string) error {
	for _, allowedMergeStatus := range allowedMergeStatuses {
		if mergeStatus == allowedMergeStatus {
			return nil
		}
	}
	return resources.InvalidVerificationMergeStatusError{"Unexpected merge status: " + mergeStatus}
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

func (verifier *Verifier) verifyRepositoryExists(repositoryId uint64) error {
	query := "SELECT id FROM repositories WHERE id=$1"
	err := verifier.database.QueryRow(query, repositoryId).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find repository with id: %d", repositoryId)
		return resources.NoSuchRepositoryError{errorText}
	}
	return nil
}

func (verifier *Verifier) verifyChangesetExists(changesetId uint64) error {
	query := "SELECT id FROM changesets WHERE id=$1"
	err := verifier.database.QueryRow(query, changesetId).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find changeset with id: %d", changesetId)
		return resources.NoSuchChangesetError{errorText}
	}
	return nil
}

func (verifier *Verifier) verifyShaPairDoesNotExist(headSha, baseSha string) error {
	query := "SELECT id FROM changesets WHERE head_sha=$1 AND base_sha=$2"
	err := verifier.database.QueryRow(query, headSha, baseSha).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := fmt.Sprintf("Changeset already exists with head sha %s and base sha %s", headSha, baseSha)
		return resources.ChangesetAlreadyExistsError{errorText}
	}
	return nil
}
