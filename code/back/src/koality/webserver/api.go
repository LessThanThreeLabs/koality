package webserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"koality/resources"
	"net/http"
	"strconv"
)

func handleApiSubroute(subrouter *mux.Router, resourcesConnection *resources.Connection) {
	subrouter.HandleFunc("/changes/create", CreateChangeHandler(resourcesConnection))
}

func CreateChangeHandler(resourcesConnection *resources.Connection) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		queryValues := request.URL.Query()

		repositoryId, err := strconv.ParseUint(queryValues.Get("repository"), 10, 64)
		if err != nil {
			fmt.Fprintf(writer, "Unable to parse query parameters")
			return
		}

		headSha := queryValues.Get("headSha")

		fmt.Fprintf(writer, "Need to create a change with repository %v and sha %v", repositoryId, headSha)
	}
}
