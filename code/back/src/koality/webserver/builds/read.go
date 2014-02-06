package builds

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (buildsHandler *BuildsHandler) get(writer http.ResponseWriter, request *http.Request) {
	buildIdString := mux.Vars(request)["buildId"]
	buildId, err := strconv.ParseUint(buildIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse buildId: %v", err)
		return
	}

	build, err := buildsHandler.resourcesConnection.Builds.Read.Get(buildId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedBuild := getSanitizedBuild(build)
	jsonedBuild, err := json.Marshal(sanitizedBuild)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedBuild)
}

func (buildsHandler *BuildsHandler) getTail(writer http.ResponseWriter, request *http.Request) {
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

	builds, err := buildsHandler.resourcesConnection.Builds.Read.GetTail(repositoryId, uint32(offset), uint32(results))
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedBuilds := make([]sanitizedBuild, 0, len(builds))
	for _, build := range builds {
		sanitizedBuilds = append(sanitizedBuilds, *getSanitizedBuild(&build))
	}

	jsonedBuilds, err := json.Marshal(sanitizedBuilds)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedBuilds)
}
