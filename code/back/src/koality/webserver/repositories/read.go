package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (repositoriesHandler *RepositoriesHandler) Get(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Read.Get(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedRepository := getSanitizedRepository(repository)
	jsonedRepository, err := json.Marshal(sanitizedRepository)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedRepository)
}

func (repositoriesHandler *RepositoriesHandler) GetAll(writer http.ResponseWriter, request *http.Request) {
	repositories, err := repositoriesHandler.resourcesConnection.Repositories.Read.GetAll()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedRepositories := make([]sanitizedRepository, 0, len(repositories))
	for _, repository := range repositories {
		sanitizedRepositories = append(sanitizedRepositories, *getSanitizedRepository(&repository))
	}

	jsonedRepositories, err := json.Marshal(sanitizedRepositories)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedRepositories)
}
