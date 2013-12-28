package stages

import (
	"koality/resources"
	"time"
)

type SubscriptionHandler struct {
	createdSubscriptionManager           resources.SubscriptionManager
	runCreatedSubscriptionManager        resources.SubscriptionManager
	returnCodeUpdatedSubscriptionManager resources.SubscriptionManager
	startTimeUpdatedSubscriptionManager  resources.SubscriptionManager
	endTimeUpdatedSubscriptionManager    resources.SubscriptionManager
	consoleLinesAddedSubscriptionManager resources.SubscriptionManager
	xunitResultsAddedSubscriptionManager resources.SubscriptionManager
	exportsAddedSubscriptionManager      resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalStagesSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToCreatedEvents(updateHandler resources.StageCreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.createdSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.createdSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCreatedEvent(stageId uint64) {
	subscriptionHandler.createdSubscriptionManager.Fire(stageId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToRunCreatedEvents(updateHandler resources.StageRunCreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.runCreatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromRunCreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.runCreatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireRunCreatedEvent(stageRunId uint64) {
	subscriptionHandler.runCreatedSubscriptionManager.Fire(stageRunId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToReturnCodeUpdatedEvents(updateHandler resources.StageReturnCodeUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.returnCodeUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromReturnCodeUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.returnCodeUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireReturnCodeUpdatedEvent(stageRunId uint64, returnCode int) {
	subscriptionHandler.returnCodeUpdatedSubscriptionManager.Fire(stageRunId, returnCode)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToStartTimeUpdatedEvents(updateHandler resources.StageStartTimeUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.startTimeUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromStartTimeUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.startTimeUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireStartTimeUpdatedEvent(stageRunId uint64, startTime time.Time) {
	subscriptionHandler.startTimeUpdatedSubscriptionManager.Fire(stageRunId, startTime)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToEndTimeUpdatedEvents(updateHandler resources.StageEndTimeUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.endTimeUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromEndTimeUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.endTimeUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireEndTimeUpdatedEvent(stageRunId uint64, endTime time.Time) {
	subscriptionHandler.endTimeUpdatedSubscriptionManager.Fire(stageRunId, endTime)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToConsoleLinesAddedEvents(updateHandler resources.StageConsoleLinesAddedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.consoleLinesAddedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromConsoleLinesAddedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.consoleLinesAddedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireConsoleLinesAddedEvent(stageRunId uint64, consoleLines map[uint64]string) {
	subscriptionHandler.consoleLinesAddedSubscriptionManager.Fire(stageRunId, consoleLines)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToXunitResultsAddedEvents(updateHandler resources.StageXunitResultsAddedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.xunitResultsAddedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromXunitResultsAddedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.xunitResultsAddedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireXunitResultsAddedEvent(stageRunId uint64, xunitResults []resources.XunitResult) {
	subscriptionHandler.xunitResultsAddedSubscriptionManager.Fire(stageRunId, xunitResults)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToExportsAddedEvents(updateHandler resources.StageExportsAddedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.exportsAddedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromExportsAddedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.exportsAddedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireExportsAddedEvent(stageRunId uint64, exports []resources.Export) {
	subscriptionHandler.exportsAddedSubscriptionManager.Fire(stageRunId, exports)
}
