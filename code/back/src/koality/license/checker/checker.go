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
	CheckLicense(licenseKey, serverId string) (*license.CheckResponse, error)
}

type licenseChecker struct {
	serverUri string
}

func New(serverUri string) LicenseChecker {
	return &licenseChecker{serverUri}
}

func (licenseChecker *licenseChecker) CheckLicense(licenseKey, serverId string) (*license.CheckResponse, error) {
	checkRequest := &license.CheckRequest{licenseKey, serverId}
	jsonedCheckRequest, err := json.Marshal(checkRequest)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(licenseChecker.serverUri+license.CheckRoute, "application/json", bytes.NewReader(jsonedCheckRequest))
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		errorMessage, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		return nil, fmt.Errorf("License check failed with HTTP status: %s, %s", response.Status, errorMessage)
	}

	checkResponse := new(license.CheckResponse)
	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(checkResponse); err != nil {
		return nil, err
	}

	return checkResponse, nil
}
