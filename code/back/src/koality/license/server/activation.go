package licenseserver

import (
	"encoding/json"
	"fmt"
	"koality/license"
	"net/http"
)

func (licenseServer *LicenseServer) deactivateLicense(writer http.ResponseWriter, request *http.Request) {
	licenseServer.handleLicenseActivationRequest(false, writer, request)
}

func (licenseServer *LicenseServer) reactivateLicense(writer http.ResponseWriter, request *http.Request) {
	licenseServer.handleLicenseActivationRequest(true, writer, request)
}

func (licenseServer *LicenseServer) handleLicenseActivationRequest(shouldBeActive bool, writer http.ResponseWriter, request *http.Request) {
	licenseActivationRequest := new(license.ActivationRequest)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(licenseActivationRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if licenseActivationRequest.LicenseKey == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a license key")
		return
	}

	query := "UPDATE licenses SET is_active=$2 WHERE key=$1"
	result, err := licenseServer.database.Exec(query, licenseActivationRequest.LicenseKey, shouldBeActive)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	count, err := result.RowsAffected()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	} else if count != 1 {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(writer, "No license found with key %s", licenseActivationRequest.LicenseKey)
		return
	}

	fmt.Fprint(writer, "ok")
}
