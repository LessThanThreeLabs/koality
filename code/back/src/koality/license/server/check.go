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
	licenseCheckRequest := new(license.CheckRequest)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(licenseCheckRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if licenseCheckRequest.LicenseKey == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a license key")
		return
	} else if licenseCheckRequest.ServerId == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a server id")
		return
	}

	licenseCheckResponse, err := licenseServer.checkLicenseKey(licenseCheckRequest)
	if err == license.NotFoundError || err == license.DeactivatedError || err == license.ServerIdMismatchError {
		licenseCheckResponse = &license.CheckResponse{
			IsValid:     false,
			ErrorReason: err.Error(),
		}
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	jsonedLicenseCheckResponse, err := json.Marshal(licenseCheckResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedLicenseCheckResponse)
}

func (licenseServer *LicenseServer) checkLicenseKey(licenseCheckRequest *license.CheckRequest) (*license.CheckResponse, error) {
	transaction, err := licenseServer.database.Begin()
	if err != nil {
		return nil, err
	}

	var maxExecutors uint32
	var isActive bool
	var serverId sql.NullString
	var lastPing *time.Time

	checkQuery := "SELECT max_executors, is_active, server_id, last_ping FROM licenses WHERE key=$1"
	err = transaction.QueryRow(checkQuery, licenseCheckRequest.LicenseKey).Scan(&maxExecutors, &isActive, &serverId, &lastPing)
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

	if serverId.String != licenseCheckRequest.ServerId && lastPing != nil && lastPing.Add(time.Hour).After(time.Now()) {
		return nil, license.ServerIdMismatchError
	}

	updateQuery := "UPDATE licenses SET server_id=$1, last_ping=$2 WHERE key=$3"
	_, err = transaction.Exec(updateQuery, licenseCheckRequest.ServerId, time.Now(), licenseCheckRequest.LicenseKey)
	if err != nil {
		return nil, err
	}

	transaction.Commit()
	licenseCheckResponse := license.CheckResponse{
		IsValid:      true,
		MaxExecutors: maxExecutors,
	}
	return &licenseCheckResponse, nil
}
