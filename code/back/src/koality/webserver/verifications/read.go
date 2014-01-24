package verifications

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (verificationsHandler *VerificationsHandler) Get(writer http.ResponseWriter, request *http.Request) {
	verificationIdString := mux.Vars(request)["verificationId"]
	verificationId, err := strconv.ParseUint(verificationIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	verification, err := verificationsHandler.resourcesConnection.Verifications.Read.Get(verificationId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerification := getSanitizedVerification(verification)
	jsonedVerification, err := json.Marshal(sanitizedVerification)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedVerification)
}

func (verificationsHandler *VerificationsHandler) GetTail(writer http.ResponseWriter, request *http.Request) {
	queryValues := request.URL.Query()

	repositoryIdString := queryValues.Get("repositoryId")
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	offsetString := queryValues.Get("offset")
	offset, err := strconv.ParseUint(offsetString, 10, 32)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	resultsString := queryValues.Get("results")
	results, err := strconv.ParseUint(resultsString, 10, 32)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	verifications, err := verificationsHandler.resourcesConnection.Verifications.Read.GetTail(repositoryId, uint32(offset), uint32(results))
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerifications := make([]sanitizedVerification, 0, len(verifications))
	for _, verification := range verifications {
		sanitizedVerifications = append(sanitizedVerifications, *getSanitizedVerification(&verification))
	}

	jsonedVerifications, err := json.Marshal(sanitizedVerifications)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedVerifications)
}
