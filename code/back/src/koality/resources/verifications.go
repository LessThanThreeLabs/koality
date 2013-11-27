package resources

import (
	"time"
)

type Verification struct {
	Id                 uint64
	RepositoryId       uint64
	MergeTarget        string
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
	Read   VerificationsReadHandler
	Update VerificationsUpdateHandler
	Delete VerificationsDeleteHandler
}

type VerificationsCreateHandler interface {
}

type VerificationsReadHandler interface {
}

type VerificationsUpdateHandler interface {
}

type VerificationsDeleteHandler interface {
}

type VerificationAlreadyExistsError struct {
	error
}

type NoSuchVerificationError struct {
	error
}
