package repositories

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (repositoriesHandler *RepositoriesHandler) delete(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	err = repositoriesHandler.resourcesConnection.Repositories.Delete.Delete(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}
