package resources

import (
	"time"
)

type Verification struct {
	Id                 uint64
	RepositoryId       uint64
	MergeTarget        string
	EmailToNotify      string
	VerificationStatus string
	MergeStatus        string
	Created            time.Time
	Started            time.Time
	Ended              time.Time
	ChangeSet          ChangeSet
}

type ChangeSet struct {
	Id           uint64
	RepositoryId uint64
	HeadSha      string
	BaseSha      string
	HeadMessage  string
	HeadUsername string
	HeadEmail    string
	Created      time.Time
}

type VerificationsHandler struct {
	Create VerificationsCreateHandler
	// Read   VerificationsReadHandler
	// Update VerificationsUpdateHandler
	// Delete VerificationsDeleteHandler
}

type VerificationsCreateHandler interface {
	Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (uint64, error)
	CreateFromChangeset(repositoryId, changesetId uint64, mergeTarget, emailToNotify string) (uint64, error)
}

type VerificationsReadHandler interface {
}

type VerificationsUpdateHandler interface {
}

type VerificationsDeleteHandler interface {
}

type NoSuchVerificationError struct {
	error
}

type ChangesetAlreadyExistsError struct {
	error
}

type NoSuchChangesetError struct {
	error
}
