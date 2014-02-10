package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"net/http"
)

func (repositoriesHandler *RepositoriesHandler) create(writer http.ResponseWriter, request *http.Request) {
	createRequestData := new(createRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(createRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Create.Create(createRequestData.Name, createRequestData.VcsType, createRequestData.RemoteUri)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = repositoriesHandler.repositoryManager.CreateRepository(repository)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)

		err = repositoriesHandler.resourcesConnection.Repositories.Delete.Delete(repository.Id)
		if err != nil {
			fmt.Fprint(writer, "\n", err)
		}
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

func (repositoriesHandler *RepositoriesHandler) createWithGitHub(writer http.ResponseWriter, request *http.Request) {
	createWithGitHubRequestData := new(createWithGitHubRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(createWithGitHubRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	userId := context.Get(request, "userId").(uint64)
	user, err := repositoriesHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		fmt.Fprint(writer, err)
		return
	} else if user.GitHubOAuth == "" && false {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, must be connected to GitHub")
		return
	}

	remoteUri := fmt.Sprintf("git@github.com:%s/%s.git", createWithGitHubRequestData.Owner, createWithGitHubRequestData.Name)
	repository, err := repositoriesHandler.resourcesConnection.Repositories.Create.CreateWithGitHub(createWithGitHubRequestData.Name, remoteUri, createWithGitHubRequestData.Owner, createWithGitHubRequestData.Name, user.GitHubOAuth)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = repositoriesHandler.repositoryManager.CreateRepository(repository)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)

		err = repositoriesHandler.resourcesConnection.Repositories.Delete.Delete(repository.Id)
		if err != nil {
			fmt.Fprint(writer, "\n", err)
		}
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
