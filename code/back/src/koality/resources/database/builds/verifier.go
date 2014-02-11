package builds

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/LessThanThreeLabs/go.codereview/patch"
	"koality/resources"
	"regexp"
	"time"
)

const (
	headShaLength          = 40
	baseShaLength          = 40
	maxHeadUsernameLength  = 128
	maxHeadEmailLength     = 256
	maxRefLength           = 1024
	maxEmailToNotifyLength = 128
	minResults             = 1
	headShaRegex           = "^[a-fA-F0-9]+$"
	baseShaRegex           = "^[a-fA-F0-9]+$"
	emailToNotifyRegex     = "^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$"
)

var (
	allowedStatuses      []string = []string{"queued", "running", "passed", "failed", "cancelled"}
	allowedMergeStatuses []string = []string{"running", "passed", "failed"}
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifySnapshotExists(snapshotId uint64) error {
	query := "SELECT id FROM snapshots WHERE id=$1"
	err := verifier.database.QueryRow(query, snapshotId).Scan(new(uint64))
	if err == sql.ErrNoRows {
		return resources.NoSuchSnapshotError{fmt.Sprintf("Snapshot with id %d does not exist.", snapshotId)}
	} else if err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyHeadSha(headSha string) error {
	if len(headSha) != headShaLength {
		return fmt.Errorf("Head SHA must be %d characters long", headShaLength)
	} else if ok, err := regexp.MatchString(headShaRegex, headSha); !ok || err != nil {
		return errors.New("Head SHA must match regex: " + headSha)
	}
	return nil
}

func (verifier *Verifier) verifyBaseSha(baseSha string) error {
	if len(baseSha) != baseShaLength {
		return fmt.Errorf("Base SHA must be %d characters long", baseShaLength)
	} else if ok, err := regexp.MatchString(baseShaRegex, baseSha); !ok || err != nil {
		return errors.New("Base SHA must match regex: " + baseSha)
	}
	return nil
}

func (verifier *Verifier) verifyHeadMessage(message string) error {
	return nil
}

func (verifier *Verifier) verifyHeadUsername(username string) error {
	if len(username) > maxHeadUsernameLength {
		return fmt.Errorf("Head username cannot exceed %d characters long", maxHeadUsernameLength)
	}
	return nil
}

func (verifier *Verifier) verifyHeadEmail(email string) error {
	if len(email) > maxHeadEmailLength {
		return fmt.Errorf("Head email cannot exceed %d characters long", maxHeadEmailLength)
	}
	return nil
}

func (verifier *Verifier) verifyPatchContents(patchContents []byte) error {
	if len(patchContents) == 0 {
		return nil
	}
	_, err := patch.Parse(patchContents)
	return err
}

func (verifier *Verifier) verifyRef(ref string) error {
	if len(ref) > maxRefLength {
		return fmt.Errorf("Merge target cannot exceed %d characters long", maxRefLength)
	}
	return nil
}

func (verifier *Verifier) verifyEmailToNotify(emailToNotify string) error {
	if len(emailToNotify) > maxEmailToNotifyLength {
		return fmt.Errorf("Email to notify cannot exceed %d characters long", maxEmailToNotifyLength)
	} else if ok, err := regexp.MatchString(emailToNotifyRegex, emailToNotify); !ok || err != nil {
		return errors.New("Email to notify must match regex: " + emailToNotifyRegex)
	}
	return nil
}

func (verifier *Verifier) verifyStatus(status string) error {
	for _, allowedBuildStatus := range allowedStatuses {
		if status == allowedBuildStatus {
			return nil
		}
	}
	return resources.InvalidBuildStatusError{"Unexpected build status: " + status}
}

func (verifier *Verifier) verifyMergeStatus(mergeStatus string) error {
	for _, allowedMergeStatus := range allowedMergeStatuses {
		if mergeStatus == allowedMergeStatus {
			return nil
		}
	}
	return resources.InvalidBuildMergeStatusError{"Unexpected merge status: " + mergeStatus}
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

func (verifier *Verifier) verifyTailResultsNumber(results uint32) error {
	if results < uint32(minResults) {
		return fmt.Errorf("Must request at least %d results", minResults)
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
