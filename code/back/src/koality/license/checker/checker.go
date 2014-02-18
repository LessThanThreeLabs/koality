package licensechecker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"koality/license"
	"net/http"
)

type LicenseChecker interface {
	CheckLicense(licenseKey, serverId string) (*license.CheckLicenseResponse, error)
}

type licenseChecker struct {
	serverUri string
}

func New(serverUri string) LicenseChecker {
	return &licenseChecker{serverUri}
}

func (licenseChecker *licenseChecker) CheckLicense(licenseKey, serverId string) (*license.CheckLicenseResponse, error) {
	checkLicenseRequest := &license.CheckLicenseRequest{licenseKey, serverId}
	jsonedCheckLicenseRequest, err := json.Marshal(checkLicenseRequest)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(licenseChecker.serverUri+license.LicenseRoute+license.CheckLicenseSubroute, "application/json", bytes.NewReader(jsonedCheckLicenseRequest))
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
