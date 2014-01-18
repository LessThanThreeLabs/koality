package snapshots

import (
	"koality/resources"
	"time"
)

type SubscriptionHandler struct {
	createdSubscriptionManager          resources.SubscriptionManager
	deletedSubscriptionManager          resources.SubscriptionManager
	statusUpdatedSubscriptionManager    resources.SubscriptionManager
	startTimeUpdatedSubscriptionManager resources.SubscriptionManager
	endTimeUpdatedSubscriptionManager   resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalSnapshotsSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToCreatedEvents(updateHandler resources.SnapshotCreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.createdSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.createdSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCreatedEvent(snapshot *resources.Snapshot) {
	subscriptionHandler.createdSubscriptionManager.Fire(snapshot)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToDeletedEvents(updateHandler resources.SnapshotDeletedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.deletedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromDeletedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.deletedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireDeletedEvent(snapshotId uint64) {
	subscriptionHandler.deletedSubscriptionManager.Fire(snapshotId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToStatusUpdatedEvents(updateHandler resources.SnapshotStatusUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.statusUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromStatusUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.statusUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireStatusUpdatedEvent(snapshotId uint64, status string) {
	subscriptionHandler.statusUpdatedSubscriptionManager.Fire(snapshotId, status)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToStartTimeUpdatedEvents(updateHandler resources.SnapshotStartTimeUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.startTimeUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromStartTimeUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.startTimeUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireStartTimeUpdatedEvent(snapshotId uint64, startTime time.Time) {
	subscriptionHandler.startTimeUpdatedSubscriptionManager.Fire(snapshotId, startTime)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToEndTimeUpdatedEvents(updateHandler resources.SnapshotEndTimeUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.endTimeUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromEndTimeUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.endTimeUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireEndTimeUpdatedEvent(snapshotId uint64, endTime time.Time) {
	subscriptionHandler.endTimeUpdatedSubscriptionManager.Fire(snapshotId, endTime)
}
