package users

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (usersHandler *UsersHandler) SetName(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	firstName := request.PostFormValue("firstName")
	lastName := request.PostFormValue("lastName")
	err := usersHandler.resourcesConnection.Users.Update.SetName(userId, firstName, lastName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (usersHandler *UsersHandler) SetPassword(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	password := request.PostFormValue("password")
	passwordHash, passwordSalt, err := usersHandler.passwordHasher.GenerateHashAndSalt(password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = usersHandler.resourcesConnection.Users.Update.SetPassword(userId, passwordHash, passwordSalt)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (usersHandler *UsersHandler) SetAdmin(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)

	userToModifyIdString := mux.Vars(request)["userId"]
	userToModifyId, err := strconv.ParseUint(userToModifyIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	adminStatusString := request.PostFormValue("admin")
	adminStatus, err := strconv.ParseBool(adminStatusString)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if userId == userToModifyId {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, cannot modify admin status for self")
		return
	}

	err = usersHandler.resourcesConnection.Users.Update.SetAdmin(userToModifyId, adminStatus)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (usersHandler *UsersHandler) AddKey(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	name := request.PostFormValue("name")
	publicKey := request.PostFormValue("publicKey")
	keyId, err := usersHandler.resourcesConnection.Users.Update.AddKey(userId, name, publicKey)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprintf(writer, "{id:%d}", keyId)
}

func (usersHandler *UsersHandler) RemoveKey(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	keyIdString := request.PostFormValue("id")
	keyId, err := strconv.ParseUint(keyIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = usersHandler.resourcesConnection.Users.Update.RemoveKey(userId, keyId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}
