package resources

import (
	"time"
)

type Build struct {
	Id            uint64
	RepositoryId  uint64
	EmailToNotify string
	Status        string
	MergeTarget   string
	MergeStatus   string
	Created       *time.Time
	Started       *time.Time
	Ended         *time.Time
	Changeset     Changeset
}

type CoreBuildInformation struct {
	RepositoryId  uint64
	HeadSha       string
	BaseSha       string
	HeadMessage   string
	HeadUsername  string
	HeadEmail     string
	EmailToNotify string
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

type BuildsHandler struct {
	Create       BuildsCreateHandler
	Read         BuildsReadHandler
	Update       BuildsUpdateHandler
	Subscription BuildsSubscriptionHandler
}

type BuildsCreateHandler interface {
	Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (*Build, error)
	CreateForSnapshot(repositoryId, snapshotId uint64, headSha, baseSha, headMessage, headUsername, headEmail, emailToNotify string) (*Build, error)
	CreateForDebugInstance(repositoryId, debugInstanceId uint64, headSha, baseSha, headMessage, headUsername, headEmail, emailToNotify string) (*Build, error)
	CreateFromChangeset(repositoryId, changesetId uint64, mergeTarget, emailToNotify string) (*Build, error)
}

type BuildsReadHandler interface {
	Get(buildId uint64) (*Build, error)
	GetTail(repositoryId uint64, offset, results uint32) ([]Build, error)
	GetChangesetFromShas(headSha, baseSha string) (*Changeset, error)
}

type BuildsUpdateHandler interface {
	SetStatus(buildId uint64, status string) error
	SetMergeStatus(buildId uint64, mergeStatus string) error
	SetStartTime(buildId uint64, startTime time.Time) error
	SetEndTime(buildId uint64, endTime time.Time) error
}

type BuildCreatedHandler func(build *Build)
type BuildStatusUpdatedHandler func(buildId uint64, status string)
type BuildMergeStatusUpdatedHandler func(buildId uint64, mergeStatus string)
type BuildStartTimeUpdatedHandler func(buildId uint64, startTime time.Time)
type BuildEndTimeUpdatedHandler func(buildId uint64, endTime time.Time)

type BuildsSubscriptionHandler interface {
	SubscribeToCreatedEvents(updateHandler BuildCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToStatusUpdatedEvents(updateHandler BuildStatusUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStatusUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToMergeStatusUpdatedEvents(updateHandler BuildMergeStatusUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromMergeStatusUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToStartTimeUpdatedEvents(updateHandler BuildStartTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStartTimeUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToEndTimeUpdatedEvents(updateHandler BuildEndTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromEndTimeUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalBuildsSubscriptionHandler interface {
	FireCreatedEvent(build *Build)
	FireStatusUpdatedEvent(buildId uint64, status string)
	FireMergeStatusUpdatedEvent(buildId uint64, mergeStatus string)
	FireStartTimeUpdatedEvent(buildId uint64, startTime time.Time)
	FireEndTimeUpdatedEvent(buildId uint64, endTime time.Time)
	BuildsSubscriptionHandler
}
