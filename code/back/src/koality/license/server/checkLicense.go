package licenseserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"koality/license"
	"net/http"
	"time"
)

func (licenseServer *LicenseServer) checkLicense(writer http.ResponseWriter, request *http.Request) {
	checkLicenseRequest := new(license.CheckLicenseRequest)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(checkLicenseRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if checkLicenseRequest.LicenseKey == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a license key")
		return
	} else if checkLicenseRequest.ServerId == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a server id")
		return
	}

	checkLicenseResponse, err := licenseServer.checkLicenseKey(checkLicenseRequest)
	if err == license.NotFoundError || err == license.DeactivatedError || err == license.ServerIdMismatchError {
		checkLicenseResponse = &license.CheckLicenseResponse{
			IsValid:     false,
			ErrorReason: err.Error(),
		}
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	jsonedLicenseCheckResponse, err := json.Marshal(checkLicenseResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedLicenseCheckResponse)
}

func (licenseServer *LicenseServer) checkLicenseKey(checkLicenseRequest *license.CheckLicenseRequest) (*license.CheckLicenseResponse, error) {
	transaction, err := licenseServer.database.Begin()
	if err != nil {
		return nil, err
	}

	var maxExecutors uint32
	var isActive bool
	var serverId sql.NullString
	var lastPing *time.Time

	checkQuery := "SELECT max_executors, is_active, server_id, last_ping FROM licenses WHERE key=$1"
	err = transaction.QueryRow(checkQuery, checkLicenseRequest.LicenseKey).Scan(&maxExecutors, &isActive, &serverId, &lastPing)
	if err == sql.ErrNoRows {
		transaction.Rollback()
		return nil, license.NotFoundError
	} else if err != nil {
		transaction.Rollback()
		return nil, err
	}

	if !isActive {
		return nil, license.DeactivatedError
	}

	if serverId.String != checkLicenseRequest.ServerId && lastPing != nil && lastPing.Add(time.Hour).After(time.Now()) {
		return nil, license.ServerIdMismatchError
	}

	updateQuery := "UPDATE licenses SET server_id=$1, last_ping=$2 WHERE key=$3"
	_, err = transaction.Exec(updateQuery, checkLicenseRequest.ServerId, time.Now(), checkLicenseRequest.LicenseKey)
	if err != nil {
		return nil, err
	}

	transaction.Commit()
	checkLicenseResponse := license.CheckLicenseResponse{
		IsValid:      true,
		MaxExecutors: maxExecutors,
	}
	return &checkLicenseResponse, nil
}
