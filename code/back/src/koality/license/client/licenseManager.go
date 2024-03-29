package client

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"koality/constants"
	"koality/license"
	"koality/resources"
	"net"
	"strings"
	"time"
)

const MaxLicenseCheckFailures = 12

type LicenseManager struct {
	resourcesConnection *resources.Connection
	licenseClient       LicenseClient
}

func NewLicenseManager(resourcesConnection *resources.Connection, licenseClient LicenseClient) *LicenseManager {
	return &LicenseManager{resourcesConnection, licenseClient}
}

func (licenseManager *LicenseManager) SetLicenseKey(licenseKey string) error {
	licenseSettings, err := licenseManager.licenseKeyToSettings(licenseKey)
	if err != nil {
		return err
	}

	if !licenseSettings.IsValid {
		return errors.New(licenseSettings.InvalidReason)
	}

	if _, err = licenseManager.resourcesConnection.Settings.Update.SetLicenseSettings(
		licenseSettings.LicenseKey, licenseSettings.ServerId, licenseSettings.MaxExecutors,
		licenseSettings.LicenseCheckFailures, licenseSettings.IsValid, licenseSettings.InvalidReason); err != nil {
		return err
	}

	return nil
}

func (licenseManager *LicenseManager) CheckUpgrade() (*license.CheckUpgradeResponse, error) {
	licenseSettings, err := licenseManager.resourcesConnection.Settings.Read.GetLicenseSettings()
	if err != nil {
		return nil, err
	}

	return licenseManager.licenseClient.CheckUpgrade(licenseSettings.LicenseKey, licenseSettings.ServerId, constants.Version)
}

func (licenseManager *LicenseManager) DownloadUpgrade(upgradeVersion string) (io.ReadCloser, error) {
	licenseSettings, err := licenseManager.resourcesConnection.Settings.Read.GetLicenseSettings()
	if err != nil {
		return nil, err
	}

	return licenseManager.licenseClient.DownloadUpgrade(licenseSettings.LicenseKey, licenseSettings.ServerId, upgradeVersion)
}

func (licenseManager *LicenseManager) licenseKeyToSettings(licenseKey string) (*resources.LicenseSettings, error) {
	licenseKey = strings.ToUpper(licenseKey)
	networkInterface, err := net.InterfaceByIndex(1)
	if err != nil {
		return nil, err
	}

	macAddress := networkInterface.HardwareAddr.String()
	hash := sha1.Sum([]byte(licenseKey + "/" + macAddress))
	serverId := hex.EncodeToString(hash[:])
	checkResponse, err := licenseManager.licenseClient.CheckLicense(licenseKey, serverId)
	if err != nil {
		return nil, err
	}

	return &resources.LicenseSettings{licenseKey, serverId, checkResponse.MaxExecutors, 0, checkResponse.IsValid, checkResponse.ErrorReason}, nil
}

func (licenseManager *LicenseManager) MonitorLicense() error {
	ticker := time.NewTicker(time.Hour)
	for {
		if err := licenseManager.checkLicenseAndRecordResult(); err != nil {
			return err
		}
		<-ticker.C
	}
}

func (licenseManager *LicenseManager) checkLicenseAndRecordResult() error {
	licenseSettings, err := licenseManager.resourcesConnection.Settings.Read.GetLicenseSettings()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		return nil
	} else if err != nil {
		return err
	}

	licenseCheckResponse, err := licenseManager.licenseClient.CheckLicense(licenseSettings.LicenseKey, licenseSettings.ServerId)
	if err != nil {
		return licenseManager.recordLicenseCheckFailure("Unexpected license check error: " + err.Error())
	} else if !licenseCheckResponse.IsValid {
		return licenseManager.recordLicenseCheckFailure(licenseCheckResponse.ErrorReason)
	} else {
		return licenseManager.recordLicenseCheckSuccess(licenseCheckResponse.MaxExecutors)
	}
}

func (licenseManager *LicenseManager) recordLicenseCheckSuccess(maxExecutors uint32) error {
	licenseSettings, err := licenseManager.resourcesConnection.Settings.Read.GetLicenseSettings()
	if err != nil {
		return err
	}

	if _, err = licenseManager.resourcesConnection.Settings.Update.SetLicenseSettings(licenseSettings.LicenseKey, licenseSettings.ServerId, maxExecutors, 0, true, ""); err != nil {
		return err
	}

	return nil
}

func (licenseManager *LicenseManager) recordLicenseCheckFailure(reason string) error {
	licenseSettings, err := licenseManager.resourcesConnection.Settings.Read.GetLicenseSettings()
	if err != nil {
		return err
	}

	licenseCheckFailures := licenseSettings.LicenseCheckFailures + 1
	isValid := licenseSettings.IsValid
	maxExecutors := licenseSettings.MaxExecutors

	if licenseCheckFailures >= MaxLicenseCheckFailures {
		isValid = false
		maxExecutors = 0
	}

	if _, err = licenseManager.resourcesConnection.Settings.Update.SetLicenseSettings(licenseSettings.LicenseKey, licenseSettings.ServerId, maxExecutors, licenseCheckFailures, isValid, reason); err != nil {
		return err
	}

	return nil
}
