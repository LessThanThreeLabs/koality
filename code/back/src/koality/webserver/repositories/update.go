package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (repositoriesHandler *RepositoriesHandler) SetGitHubHookTypes(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	user, err := repositoriesHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	} else if !user.IsAdmin || user.IsDeleted {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, must be an admin")
		return
	}

	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
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

	fmt.Println("...need to actually set the hook in github")
	fmt.Println("...make sure to delete old hooks, if there are any")
	hookId := int64(17)
	hookSecret := "hook-secret"
	err = repositoriesHandler.resourcesConnection.Repositories.Update.SetGitHubHook(repositoryId, hookId, hookSecret, hookTypes)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (repositoriesHandler *RepositoriesHandler) ClearGitHubHook(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	user, err := repositoriesHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	} else if !user.IsAdmin || user.IsDeleted {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, must be an admin")
		return
	}

	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
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

	fmt.Println("...need to actually remove hooks from github")
	fmt.Fprint(writer, "ok")
}
