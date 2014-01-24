package resources

import (
	"time"
)

type Stage struct {
	Id             uint64
	VerificationId uint64
	SectionNumber  uint64
	Name           string
	OrderNumber    uint64
	Runs           []StageRun // Do we need this? Is this helpful?
}

type StageRun struct {
	Id         uint64
	StageId    uint64
	ReturnCode int // ignore if -1
	Created    *time.Time
	Started    *time.Time
	Ended      *time.Time
}

type XunitResult struct {
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Sysout      string  `json:"sysout"`
	Syserr      string  `json:"syserr"`
	FailureText string  `json:"failure"`
	ErrorText   string  `json:"error"`
	Seconds     float64 `json:"time"`
}

type Export struct {
	BucketName string
	Path       string
	Key        string
}

type StagesHandler struct {
	Create       StagesCreateHandler
	Read         StagesReadHandler
	Update       StagesUpdateHandler
	Subscription StagesSubscriptionHandler
}

type StagesCreateHandler interface {
	Create(verificationId, sectionNumber uint64, name string, orderNumber uint64) (*Stage, error)
	CreateRun(stageId uint64) (*StageRun, error)
}

type StagesReadHandler interface {
	Get(stageId uint64) (*Stage, error)
	GetBySectionNumberAndName(verificationId, sectionNumber uint64, name string) (*Stage, error)
	GetAll(verificationId uint64) ([]Stage, error)
	GetRun(stageRunId uint64) (*StageRun, error)
	GetAllRuns(stageId uint64) ([]StageRun, error)
	GetConsoleLinesHead(stageRunId uint64, offset, results uint32) (map[uint64]string, error)
	GetConsoleLinesTail(stageRunId uint64, offset, results uint32) (map[uint64]string, error)
	GetAllConsoleLines(stageRunId uint64) (map[uint64]string, error)
	GetAllXunitResults(stageRunId uint64) ([]XunitResult, error)
	GetExports(stageRunId uint64) ([]Export, error)
}

type StagesUpdateHandler interface {
	SetReturnCode(stageRunId uint64, returnCode int) error
	SetStartTime(stageRunId uint64, startTime time.Time) error
	SetEndTime(stageRunId uint64, endTime time.Time) error
	AddConsoleLines(stageRunId uint64, consoleLines map[uint64]string) error
	RemoveAllConsoleLines(stageRunId uint64) error
	AddXunitResults(stageRunId uint64, xunitResults []XunitResult) error
	RemoveAllXunitResults(stageRunId uint64) error
	AddExports(stageRunId uint64, exports []Export) error
}

type StageCreatedHandler func(stage *Stage)
type StageRunCreatedHandler func(stageRun *StageRun)
type StageReturnCodeUpdatedHandler func(stageRunId uint64, returnCode int)
type StageStartTimeUpdatedHandler func(stageRunId uint64, startTime time.Time)
type StageEndTimeUpdatedHandler func(stageRunId uint64, endTime time.Time)
type StageConsoleLinesAddedHandler func(stageRunId uint64, consoleLines map[uint64]string)
type StageXunitResultsAddedHandler func(stageRunId uint64, xunitResults []XunitResult)
type StageExportsAddedHandler func(stageRunId uint64, exports []Export)

type StagesSubscriptionHandler interface {
	SubscribeToCreatedEvents(updateHandler StageCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToRunCreatedEvents(updateHandler StageRunCreatedHandler) (SubscriptionId, error)
	UnsubscribeFromRunCreatedEvents(subscriptionId SubscriptionId) error

	SubscribeToReturnCodeUpdatedEvents(updateHandler StageReturnCodeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromReturnCodeUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToStartTimeUpdatedEvents(updateHandler StageStartTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromStartTimeUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToEndTimeUpdatedEvents(updateHandler StageEndTimeUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromEndTimeUpdatedEvents(subscriptionId SubscriptionId) error

	SubscribeToConsoleLinesAddedEvents(updateHandler StageConsoleLinesAddedHandler) (SubscriptionId, error)
	UnsubscribeFromConsoleLinesAddedEvents(subscriptionId SubscriptionId) error

	SubscribeToXunitResultsAddedEvents(updateHandler StageXunitResultsAddedHandler) (SubscriptionId, error)
	UnsubscribeFromXunitResultsAddedEvents(subscriptionId SubscriptionId) error

	SubscribeToExportsAddedEvents(updateHandler StageExportsAddedHandler) (SubscriptionId, error)
	UnsubscribeFromExportsAddedEvents(subscriptionId SubscriptionId) error
}

type InternalStagesSubscriptionHandler interface {
	FireCreatedEvent(stage *Stage)
	FireRunCreatedEvent(stageRun *StageRun)
	FireReturnCodeUpdatedEvent(stageRunId uint64, returnCode int)
	FireStartTimeUpdatedEvent(stageRunId uint64, startTime time.Time)
	FireEndTimeUpdatedEvent(stageRunId uint64, endTime time.Time)
	FireConsoleLinesAddedEvent(stageRunId uint64, consoleLines map[uint64]string)
	FireXunitResultsAddedEvent(stageRunId uint64, xunitResuts []XunitResult)
	FireExportsAddedEvent(stageRunId uint64, exports []Export)
	StagesSubscriptionHandler
}
