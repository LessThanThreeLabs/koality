package repositories

import (
	"fmt"
	"github.com/gorilla/context"
	"net/http"
)

func (repositoriesHandler *RepositoriesHandler) Create(writer http.ResponseWriter, request *http.Request) {
	name := request.PostFormValue("name")
	vcsType := request.PostFormValue("vcsType")
	remoteUri := request.PostFormValue("remoteUri")

	fmt.Println("...need to actually get the local uri here")
	localUri := "git/local_uri/name.git"

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Create.Create(name, vcsType, localUri, remoteUri)
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
	fmt.Fprintf(writer, "{id:%d}", repository.Id)
}

func (repositoriesHandler *RepositoriesHandler) CreateWithGitHub(writer http.ResponseWriter, request *http.Request) {
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

	fmt.Println("...need to actually get the local and remote uri here")
	localUri := "git/local_uri/name.git"
	remoteUri := "git@remote.uri.com:name.git"

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Create.CreateWithGitHub(name, localUri, remoteUri, owner, name)
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
	fmt.Fprintf(writer, "{id:%d}", repository.Id)
}
