package license

import "errors"

// License Check

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

type CheckLicenseRequest struct {
	LicenseKey string `json:"licenseKey"`
	ServerId   string `json:"serverId"`
}

type CheckLicenseResponse struct {
	IsValid      bool   `json:"isValid"`
	ErrorReason  string `json:"errorReason,omitempty"`
	MaxExecutors uint32 `json:"maxExecutors,omitempty"`
}

// License Activation

type LicenseActivationRequest struct {
	LicenseKey string `json:"licenseKey"`
}

// Set License Max Executors

type SetMaxExecutorsRequest struct {
	LicenseKey   string `json:"licenseKey"`
	MaxExecutors uint32 `json:"maxExecutors"`
}

// Generate License

type GenerateLicenseRequest struct {
	MaxExecutors uint32 `json:"maxExecutors"`
}

type GenerateLicenseResponse struct {
	LicenseKey string `json:"licenseKey"`
}

// Check Upgrade

type CheckUpgradeRequest struct {
	LicenseKey     string `json:"licenseKey"`
	ServerId       string `json:"serverId"`
	CurrentVersion string `json:"currentVersion"`
}

type CheckUpgradeResponse struct {
	HasUpgrade bool      `json:"hasUpgrade"`
	NewVersion string    `json:"newVersion,omitempty"`
	Changelog  Changelog `json:"changelog,omitempty"`
}

type Changelog []ChangeInfo

type ChangeInfo struct {
	VersionAdded string   `json:"versionAdded"`
	Changes      []string `json:"changes"`
}

// Download Upgrade

type DownloadUpgradeRequest struct {
	LicenseKey string `json:"licenseKey"`
	ServerId   string `json:"serverId"`
	Version    string `json:"version"`
}
