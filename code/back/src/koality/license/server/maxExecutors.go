package licenseserver

import (
	"encoding/json"
	"fmt"
	"koality/license"
	"net/http"
)

func (licenseServer *LicenseServer) setMaxExecutors(writer http.ResponseWriter, request *http.Request) {
	setMaxExecutorsRequest := new(license.SetMaxExecutorsRequest)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(setMaxExecutorsRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if setMaxExecutorsRequest.LicenseKey == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a license key")
		return
	} else if setMaxExecutorsRequest.MaxExecutors == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a positive value for the new max executors limit")
		return
	}

	query := "UPDATE licenses SET max_executors=$2 WHERE key=$1"
	result, err := licenseServer.database.Exec(query, setMaxExecutorsRequest.LicenseKey, setMaxExecutorsRequest.MaxExecutors)
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
		fmt.Fprintf(writer, "No license found with key %s", setMaxExecutorsRequest.LicenseKey)
		return
	}

	fmt.Fprint(writer, "ok")
}
