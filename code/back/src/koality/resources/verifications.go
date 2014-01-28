package resources

import (
	"time"
)

type Verification struct {
	Id            uint64
	RepositoryId  uint64
	SnapshotId    uint64
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
	Create       VerificationsCreateHandler
	Read         VerificationsReadHandler
	Update       VerificationsUpdateHandler
	Subscription VerificationsSubscriptionHandler
}

type VerificationsCreateHandler interface {
	Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (*Verification, error)
	CreateForSnapshot(repositoryId, snapshotId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (*Verification, error)
	CreateFromChangeset(repositoryId, changesetId uint64, mergeTarget, emailToNotify string) (*Verification, error)
}

type VerificationsReadHandler interface {
	Get(verificationId uint64) (*Verification, error)
	GetTail(repositoryId uint64, offset, results uint32) ([]Verification, error)
	// GetChangesetFromShas(headSha, baseSha string) (*Changeset, error)
}

type VerificationsUpdateHandler interface {
	SetStatus(verificationId uint64, status string) error
	SetMergeStatus(verificationId uint64, mergeStatus string) error
	SetStartTime(verificationId uint64, startTime time.Time) error
	SetEndTime(verificationId uint64, endTime time.Time) error
}

type VerificationCreatedHandler func(verification *Verification)
type VerificationStatusUpdatedHandler func(verificationId uint64, status string)
type VerificationMergeStatusUpdatedHandler func(verificationId uint64, mergeStatus string)
type VerificationStartTimeUpdatedHandler func(verificationId uint64, startTime time.Time)
type VerificationEndTimeUpdatedHandler func(verificationId uint64, endTime time.Time)

type VerificationsSubscriptionHandler interface {
	SubscribeToCreatedEvents(updateHandler VerificationCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToStatusUpdatedEvents(updateHandler VerificationStatusUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStatusUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToMergeStatusUpdatedEvents(updateHandler VerificationMergeStatusUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromMergeStatusUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToStartTimeUpdatedEvents(updateHandler VerificationStartTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStartTimeUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToEndTimeUpdatedEvents(updateHandler VerificationEndTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromEndTimeUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalVerificationsSubscriptionHandler interface {
	FireCreatedEvent(verification *Verification)
	FireStatusUpdatedEvent(verificationId uint64, status string)
	FireMergeStatusUpdatedEvent(verificationId uint64, mergeStatus string)
	FireStartTimeUpdatedEvent(verificationId uint64, startTime time.Time)
	FireEndTimeUpdatedEvent(verificationId uint64, endTime time.Time)
	VerificationsSubscriptionHandler
}
