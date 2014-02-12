package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"koality/github"
	"net/http"
	"strconv"
)

func (repositoriesHandler *RepositoriesHandler) get(writer http.ResponseWriter, request *http.Request) {
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

func (repositoriesHandler *RepositoriesHandler) getAll(writer http.ResponseWriter, request *http.Request) {
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

func (repositoriesHandler *RepositoriesHandler) getGitHubRepositories(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	user, err := repositoriesHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	repositories, err := repositoriesHandler.gitHubConnection.GetRepositories(user.GitHubOAuth)
	if _, ok := err.(github.InvalidOAuthTokenError); ok {
		authorizationUrl, err := repositoriesHandler.gitHubConnection.GetAuthorizationUrl("addRepository")
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
			return
		}
		writer.WriteHeader(http.StatusPreconditionFailed)
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, fmt.Sprintf(`{"redirectUri": "%s"}`, authorizationUrl))
		return
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	jsonedRepositories, err := json.Marshal(repositories)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedRepositories)
}
