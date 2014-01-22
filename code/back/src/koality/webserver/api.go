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
			fmt.Fprintf(writer, "Unable to parse query parameters")
			return
		}

		headSha := queryValues.Get("headSha")

		repository, err := resourcesConnection.Repositories.Read.Get(repositoryId)
		if err != nil {
			fmt.Fprintf(writer, "I shit my pants")
		}

		repositoryManager.StorePending(repository, headSha)

		message, username, email, err := repositoryManager.GetCommitAttributes(repository, headSha)
		if err != nil {
			fmt.Fprintf(writer, "I shit my pants")
		}

		fmt.Fprintf(writer, "Need to create a change with repository %v, message %s, username %s and email %s ", repository, message, username, email)
	}
}
