package users

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (usersHandler *UsersHandler) delete(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)

	userToDeleteIdString := mux.Vars(request)["userId"]
	userToDeleteId, err := strconv.ParseUint(userToDeleteIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse userId: %v", err)
		return
	}

	if userId == userToDeleteId {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, cannot delete self")
		return
	}

	// Even though we will mark the user as deleted, the user is still retrievable.
	// We disable admin just to be extra safe
	err = usersHandler.resourcesConnection.Users.Update.SetAdmin(userToDeleteId, false)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = usersHandler.resourcesConnection.Users.Delete.Delete(userToDeleteId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}
