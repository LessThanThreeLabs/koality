package resources

import (
	"time"
)

type Snapshot struct {
	Id        uint64
	PoolId    uint64
	ImageId   string
	Status    string
	Created   *time.Time
	Started   *time.Time
	Ended     *time.Time
	IsDeleted bool
}

type SnapshotsHandler struct {
	Create       SnapshotsCreateHandler
	Read         SnapshotsReadHandler
	Update       SnapshotsUpdateHandler
	Delete       SnapshotsDeleteHandler
	Subscription SnapshotsSubscriptionHandler
}

type SnapshotsCreateHandler interface {
	Create(poolId uint64, repositoryInformation []*CoreBuildInformation) (*Snapshot, error)
}

type SnapshotsReadHandler interface {
	Get(snapshotId uint64) (*Snapshot, error)
	GetByImageId(imageId string) (*Snapshot, error)
	GetAllForPool(poolId uint64) ([]Snapshot, error)
}

type SnapshotsUpdateHandler interface {
	SetStatus(snapshotId uint64, status string) error
	SetImageId(snapshotId uint64, imageId string) error
	SetStartTime(snapshotId uint64, startTime time.Time) error
	SetEndTime(snapshotId uint64, endTime time.Time) error
}

type SnapshotsDeleteHandler interface {
	Delete(snapshotId uint64) error
}

type SnapshotCreatedHandler func(Snapshot *Snapshot)
type SnapshotDeletedHandler func(ec2PoolId uint64)
type SnapshotStatusUpdatedHandler func(snapshotId uint64, status string)
type SnapshotImageIdUpdatedHandler func(snapshotId uint64, imageId string)
type SnapshotStartTimeUpdatedHandler func(snapshotId uint64, startTime time.Time)
type SnapshotEndTimeUpdatedHandler func(snapshotId uint64, endTime time.Time)

type SnapshotsSubscriptionHandler interface {
	SubscribeToCreatedEvents(updateHandler SnapshotCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToDeletedEvents(updateHandler SnapshotDeletedHandler) (SubscriptionId, error)
	UnsubscribeFromDeletedEvents(subscriptionId SubscriptionId) error

	SubscribeToStatusUpdatedEvents(updateHandler SnapshotStatusUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStatusUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToImageIdUpdatedEvents(updateHandler SnapshotImageIdUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromImageIdUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToStartTimeUpdatedEvents(updateHandler SnapshotStartTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStartTimeUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToEndTimeUpdatedEvents(updateHandler SnapshotEndTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromEndTimeUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalSnapshotsSubscriptionHandler interface {
	FireCreatedEvent(snapshot *Snapshot)
	FireDeletedEvent(snapshotId uint64)
	FireStatusUpdatedEvent(snapshotId uint64, status string)
	FireImageIdUpdatedEvent(snapshotId uint64, imageId string)
	FireStartTimeUpdatedEvent(snapshotId uint64, startTime time.Time)
	FireEndTimeUpdatedEvent(snapshotId uint64, endTime time.Time)
	SnapshotsSubscriptionHandler
}
