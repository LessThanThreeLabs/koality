package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func (repositoriesHandler *RepositoriesHandler) setGitHubHookTypes(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	hookTypesString := request.PostFormValue("hookTypes")
	var hookTypes []string
	err = json.Unmarshal([]byte(hookTypesString), &hookTypes)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Read.Get(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprint(writer, err)
		return
	}

	if repository.GitHub == nil {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, repository was not added from GitHub")
		return
	}

	hookSecret := generateSecret()
	hookId, err := repositoriesHandler.gitHubConnection.AddRepositoryHook(repository, hookTypes, hookSecret)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = repositoriesHandler.resourcesConnection.Repositories.Update.SetGitHubHook(repositoryId, hookId, hookSecret, hookTypes)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if repository.GitHub.HookId != 0 {
		err = repositoriesHandler.gitHubConnection.RemoveRepositoryHook(repository)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
			return
		}
	}

	fmt.Fprint(writer, "ok")
}

func (repositoriesHandler *RepositoriesHandler) clearGitHubHook(writer http.ResponseWriter, request *http.Request) {
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

	if repository.GitHub == nil {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, repository was not added from GitHub")
		return
	}

	if repository.GitHub.HookId == 0 {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, repository does not have a GitHub hook")
		return
	}

	err = repositoriesHandler.gitHubConnection.RemoveRepositoryHook(repository)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = repositoriesHandler.resourcesConnection.Repositories.Update.ClearGitHubHook(repository.Id)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

func generateSecret() string {
	// 36^12 < 2^63; does not overflow
	maxRandValue := int64(1)
	for i := 0; i < 12; i++ {
		maxRandValue *= 36
	}
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.FormatInt(random.Int63n(maxRandValue), 36)
}
