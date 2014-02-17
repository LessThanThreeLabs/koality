package license

import "errors"

// Check

const (
	LicenseNotFound    = "License does not exist."
	LicenseDeactivated = "The license has been deactivated."
	ServerIdMismatch   = "License already in use by another instance of Koality."
)

var (
	NotFoundError         = errors.New(LicenseNotFound)
	DeactivatedError      = errors.New(LicenseDeactivated)
	ServerIdMismatchError = errors.New(ServerIdMismatch)
)

type CheckRequest struct {
	LicenseKey string `json:"licenseKey"`
	ServerId   string `json:"serverId"`
}

type CheckResponse struct {
	IsValid      bool   `json:"isValid"`
	ErrorReason  string `json:"errorReason,omitempty"`
	MaxExecutors uint32 `json:"maxExecutors,omitempty"`
}

// Activation

type ActivationRequest struct {
	LicenseKey string `json:"licenseKey"`
}

// Set Max Executors

type SetMaxExecutorsRequest struct {
	LicenseKey   string `json:"licenseKey"`
	MaxExecutors uint32 `json:"maxExecutors"`
}

// Generate

type GenerateRequest struct {
	MaxExecutors uint32 `json:"maxExecutors"`
}

type GenerateResponse struct {
	LicenseKey string `json:"licenseKey"`
}
