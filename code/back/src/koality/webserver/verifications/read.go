package builds

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (buildsHandler *VerificationsHandler) get(writer http.ResponseWriter, request *http.Request) {
	buildIdString := mux.Vars(request)["buildId"]
	buildId, err := strconv.ParseUint(buildIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse buildId: %v", err)
		return
	}

	build, err := buildsHandler.resourcesConnection.Verifications.Read.Get(buildId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerification := getSanitizedVerification(build)
	jsonedVerification, err := json.Marshal(sanitizedVerification)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedVerification)
}

func (buildsHandler *VerificationsHandler) getTail(writer http.ResponseWriter, request *http.Request) {
	queryValues := request.URL.Query()

	repositoryIdString := queryValues.Get("repositoryId")
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	offsetString := queryValues.Get("offset")
	offset, err := strconv.ParseUint(offsetString, 10, 32)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse offset: %v", err)
		return
	}

	resultsString := queryValues.Get("results")
	results, err := strconv.ParseUint(resultsString, 10, 32)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse results: %v", err)
		return
	}

	builds, err := buildsHandler.resourcesConnection.Verifications.Read.GetTail(repositoryId, uint32(offset), uint32(results))
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerifications := make([]sanitizedVerification, 0, len(builds))
	for _, build := range builds {
		sanitizedVerifications = append(sanitizedVerifications, *getSanitizedVerification(&build))
	}

	jsonedVerifications, err := json.Marshal(sanitizedVerifications)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedVerifications)
}
