package resources

import (
	"time"
)

type Stage struct {
	Id             uint64
	VerificationId uint64
	Name           string
	Flavor         string
	OrderNumber    uint64
	Runs           []StageRun // Do we need this? Is this helpful?
}

type StageRun struct {
	Id         uint64
	ReturnCode int // ignore if -1
	Created    *time.Time
	Started    *time.Time
	Ended      *time.Time
}

type XunitResult struct {
	Name        string
	Path        string
	Sysout      string
	Syserr      string
	FailureText string
	ErrorText   string
	Started     time.Time
	Seconds     float64
}

type Export struct {
	Path string
	Uri  string
}

type StagesHandler struct {
	Create StagesCreateHandler
	Read   StagesReadHandler
	Update StagesUpdateHandler
}

type StagesCreateHandler interface {
	Create(verificationId uint64, name, flavor string, orderNumber uint64) (uint64, error)
	CreateRun(stageId uint64) (uint64, error)
}

type StagesReadHandler interface {
	Get(stageId uint64) (*Stage, error)
	GetAll(verificationId uint64) ([]Stage, error)
	GetRun(stageRunId uint64) (*StageRun, error)
	GetAllRuns(stageId uint64) ([]StageRun, error)
	GetConsoleTextHead(stageRunId uint64, offset, results int) (map[uint64]string, error)
	GetConsoleTextTail(stageRunId uint64, offset, results int) (map[uint64]string, error)
	GetAllConsoleText(stageRunId uint64) (map[uint64]string, error)
	GetXunitResults(stageRunId uint64) ([]XunitResult, error)
	GetExports(stageRunId uint64) ([]Export, error)
}

type StagesUpdateHandler interface {
	SetReturnCode(stageRunId uint64, returnCode int) error
	SetStartTime(verificationId uint64, startTime time.Time) error
	SetEndTime(verificationId uint64, endTime time.Time) error
	AddConsoleLines(stageRunId uint64, consoleTextLines map[uint64]string) error
	AddXunitResults(stageRunId uint64, xunitResults []XunitResult) error
	AddExports(stageRunId uint64, exports []Export) error
}

type NoSuchStageError struct {
	Message string
}

func (err NoSuchStageError) Error() string {
	return err.Message
}

type StageAlreadyExistsError struct {
	Message string
}

func (err StageAlreadyExistsError) Error() string {
	return err.Message
}

type InvalidStageFlavorError struct {
	Message string
}

func (err InvalidStageFlavorError) Error() string {
	return err.Message
}

type NoSuchStageRunError struct {
	Message string
}

func (err NoSuchStageRunError) Error() string {
	return err.Message
}

type NoSuchXunitError struct {
	Message string
}

func (err NoSuchXunitError) Error() string {
	return err.Message
}
