package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"net/http"
)

func (repositoriesHandler *RepositoriesHandler) create(writer http.ResponseWriter, request *http.Request) {
	name := request.PostFormValue("name")
	vcsType := request.PostFormValue("vcsType")
	remoteUri := request.PostFormValue("remoteUri")

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Create.Create(name, vcsType, remoteUri)
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
	userId := context.Get(request, "userId").(uint64)
	user, err := repositoriesHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		fmt.Fprint(writer, err)
		return
	} else if user.GitHubOauth == "" && false {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, must be connected to GitHub")
		return
	}

	owner := request.PostFormValue("owner")
	name := request.PostFormValue("name")

	remoteUri := fmt.Sprintf("git@github.com:%s/%s.git", owner, name)

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Create.CreateWithGitHub(name, remoteUri, owner, name)
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
