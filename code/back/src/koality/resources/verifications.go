package resources

import (
	"time"
)

type Verification struct {
	Id            uint64
	RepositoryId  uint64
	MergeTarget   string
	EmailToNotify string
	Status        string
	MergeStatus   string
	Created       *time.Time
	Started       *time.Time
	Ended         *time.Time
	Changeset     Changeset
}

type Changeset struct {
	Id           uint64
	RepositoryId uint64
	HeadSha      string
	BaseSha      string
	HeadMessage  string
	HeadUsername string
	HeadEmail    string
	Created      *time.Time
}

type VerificationsHandler struct {
	Create VerificationsCreateHandler
	Read   VerificationsReadHandler
	Update VerificationsUpdateHandler
}

type VerificationsCreateHandler interface {
	Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (uint64, error)
	CreateFromChangeset(repositoryId, changesetId uint64, mergeTarget, emailToNotify string) (uint64, error)
}

type VerificationsReadHandler interface {
	Get(verificationId uint64) (*Verification, error)
	GetTail(repositoryId uint64, offset, results int) ([]Verification, error)
}

type VerificationsUpdateHandler interface {
	SetStatus(verificationId uint64, status string) error
	SetMergeStatus(verificationId uint64, mergeStatus string) error
	SetStartTime(verificationId uint64, startTime time.Time) error
	SetEndTime(verificationId uint64, endTime time.Time) error
}

type NoSuchVerificationError struct {
	Message string
}

func (err NoSuchVerificationError) Error() string {
	return err.Message
}

type InvalidVerificationStatusError struct {
	Message string
}

func (err InvalidVerificationStatusError) Error() string {
	return err.Message
}

type InvalidVerificationMergeStatusError struct {
	Message string
}

func (err InvalidVerificationMergeStatusError) Error() string {
	return err.Message
}

type ChangesetAlreadyExistsError struct {
	Message string
}

func (err ChangesetAlreadyExistsError) Error() string {
	return err.Message
}

type NoSuchChangesetError struct {
	Message string
}

func (err NoSuchChangesetError) Error() string {
	return err.Message
}
