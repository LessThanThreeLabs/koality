package verifications

import (
	"koality/resources"
	"time"
)

type SubscriptionHandler struct {
	createdSubscriptionManager            resources.SubscriptionManager
	statusUpdatedSubscriptionManager      resources.SubscriptionManager
	mergeStatusUpdatedSubscriptionManager resources.SubscriptionManager
	startTimeUpdatedSubscriptionManager   resources.SubscriptionManager
	endTimeUpdatedSubscriptionManager     resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalVerificationsSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToCreatedEvents(updateHandler resources.VerificationCreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.createdSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.createdSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCreatedEvent(verification *resources.Verification) {
	subscriptionHandler.createdSubscriptionManager.Fire(verification)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToStatusUpdatedEvents(updateHandler resources.VerificationStatusUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.statusUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromStatusUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.statusUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireStatusUpdatedEvent(verificationId uint64, status string) {
	subscriptionHandler.statusUpdatedSubscriptionManager.Fire(verificationId, status)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToMergeStatusUpdatedEvents(updateHandler resources.VerificationMergeStatusUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.mergeStatusUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromMergeStatusUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.mergeStatusUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireMergeStatusUpdatedEvent(verificationId uint64, mergeStatus string) {
	subscriptionHandler.mergeStatusUpdatedSubscriptionManager.Fire(verificationId, mergeStatus)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToStartTimeUpdatedEvents(updateHandler resources.VerificationStartTimeUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.startTimeUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromStartTimeUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.startTimeUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireStartTimeUpdatedEvent(verificationId uint64, startTime time.Time) {
	subscriptionHandler.startTimeUpdatedSubscriptionManager.Fire(verificationId, startTime)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToEndTimeUpdatedEvents(updateHandler resources.VerificationEndTimeUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.endTimeUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromEndTimeUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.endTimeUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireEndTimeUpdatedEvent(verificationId uint64, endTime time.Time) {
	subscriptionHandler.endTimeUpdatedSubscriptionManager.Fire(verificationId, endTime)
}
