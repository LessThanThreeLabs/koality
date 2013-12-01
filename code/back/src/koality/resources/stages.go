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
	Runs           []StageRun
}

type StageRun struct {
	Id         uint64
	ReturnCode uint64
	Created    *time.Time
	Started    *time.Time
	Ended      *time.Time
}

type StagesHandler struct {
	Create StagesCreateHandler
	// Read   StagesReadHandler
	// Update StagesUpdateHandler
}

type StagesCreateHandler interface {
	// Create(repositoryId uint64, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify string) (uint64, error)
	// CreateFromChangeset(repositoryId, changesetId uint64, mergeTarget, emailToNotify string) (uint64, error)
}

type StagesReadHandler interface {
	// Get(verificationId uint64) (*Verification, error)
}

type StagesUpdateHandler interface {
	// SetStatus(verificationId uint64, status string) error
	// SetMergeStatus(verificationId uint64, mergeStatus string) error
	// SetStartTime(verificationId uint64, startTime time.Time) error
	// SetEndTime(verificationId uint64, endTime time.Time) error
}

type NoSuchStageError struct {
	error
}

type StageAlreadyExistsError struct {
	error
}

type InvalidStageFlavorError struct {
	error
}

type NoSuchStageRunError struct {
	error
}
