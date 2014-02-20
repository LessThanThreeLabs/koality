package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"koality/license"
	"net/http"
)

type LicenseClient interface {
	CheckLicense(licenseKey, serverId string) (*license.CheckLicenseResponse, error)
	CheckUpgrade(licenseKey, serverId, currentVersion string) (*license.CheckUpgradeResponse, error)
	DownloadUpgrade(licenseKey, serverId, version string) (upgradeReader io.ReadCloser, err error)
}

type licenseClient struct {
	serverUri string
}

func New(serverUri string) LicenseClient {
	return &licenseClient{serverUri}
}

func (licenseClient *licenseClient) CheckLicense(licenseKey, serverId string) (*license.CheckLicenseResponse, error) {
	checkLicenseRequest := &license.CheckLicenseRequest{licenseKey, serverId}
	jsonedCheckLicenseRequest, err := json.Marshal(checkLicenseRequest)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(licenseClient.serverUri+license.LicenseRoute+license.CheckLicenseSubroute, "application/json", bytes.NewReader(jsonedCheckLicenseRequest))
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		errorMessage, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		return nil, fmt.Errorf("License check failed with HTTP status: %s, %s", response.Status, errorMessage)
	}

	checkLicenseResponse := new(license.CheckLicenseResponse)
	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(checkLicenseResponse); err != nil {
		return nil, err
	}

	return checkLicenseResponse, nil
}

func (licenseClient *licenseClient) CheckUpgrade(licenseKey, serverId, currentVersion string) (*license.CheckUpgradeResponse, error) {
	checkUpgradeRequest := &license.CheckUpgradeRequest{licenseKey, serverId, currentVersion}
	jsonedCheckUpgradeRequest, err := json.Marshal(checkUpgradeRequest)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(licenseClient.serverUri+license.UpgradeRoute+license.CheckUpgradeSubroute, "application/json", bytes.NewReader(jsonedCheckUpgradeRequest))
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		errorMessage, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		return nil, fmt.Errorf("Upgrade check failed with HTTP status: %s, %s", response.Status, errorMessage)
	}

	checkUpgradeResponse := new(license.CheckUpgradeResponse)
	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(checkUpgradeResponse); err != nil {
		return nil, err
	}

	return checkUpgradeResponse, nil
}

func (licenseClient *licenseClient) DownloadUpgrade(licenseKey, serverId, version string) (io.ReadCloser, error) {
	downloadUpgradeRequest := &license.DownloadUpgradeRequest{licenseKey, serverId, version}
	jsonedDownloadUpgradeRequest, err := json.Marshal(downloadUpgradeRequest)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(licenseClient.serverUri+license.UpgradeRoute+license.DownloadUpgradeSubroute, "application/json", bytes.NewReader(jsonedDownloadUpgradeRequest))
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		errorMessage, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		return nil, fmt.Errorf("Upgrade check failed with HTTP status: %s, %s", response.Status, errorMessage)
	}

	return response.Body, nil
}
