package webserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"koality/repositorymanager"
	"koality/resources"
	"net/http"
	"strconv"
)

func handleApiSubroute(subrouter *mux.Router, resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager) {
	subrouter.HandleFunc("/changes/create", CreateChangeHandler(resourcesConnection, repositoryManager))
}

func CreateChangeHandler(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		queryValues := request.URL.Query()

		repositoryId, err := strconv.ParseUint(queryValues.Get("repository"), 10, 64)
		if err != nil {
			fmt.Fprint(writer, err)
			return
		}

		headSha := queryValues.Get("headSha")

		repository, err := resourcesConnection.Repositories.Read.Get(repositoryId)
		if err != nil {
			fmt.Fprint(writer, err)
			return
		}

		repositoryManager.StorePending(repository, headSha)

		headMessage, headUsername, headEmail, err := repositoryManager.GetCommitAttributes(repository, headSha)
		if err != nil {
			fmt.Fprint(writer, err)
			return
		}

		verification, err := resourcesConnection.Verifications.Create.Create(repositoryId, headSha, headSha, headMessage, headUsername, headEmail, "", headEmail)
		if err != nil {
			fmt.Fprint(writer, err)
			return
		}
		fmt.Fprint(writer, verification.Id)
	}
}
