package users

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (usersHandler *UsersHandler) Get(writer http.ResponseWriter, request *http.Request) {
	userIdString := mux.Vars(request)["userId"]
	userId, err := strconv.ParseUint(userIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	user, err := usersHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedUser := getSanitizedUser(user)
	jsonedUser, err := json.Marshal(sanitizedUser)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedUser)
}

func (usersHandler *UsersHandler) GetAll(writer http.ResponseWriter, request *http.Request) {
	users, err := usersHandler.resourcesConnection.Users.Read.GetAll()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedUsers := make([]sanitizedUser, 0, 10)
	for _, user := range users {
		sanitizedUsers = append(sanitizedUsers, *getSanitizedUser(&user))
	}

	jsonedUsers, err := json.Marshal(sanitizedUsers)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedUsers)
}

func (usersHandler *UsersHandler) GetKeys(writer http.ResponseWriter, request *http.Request) {
	userIdString := mux.Vars(request)["userId"]
	userId, err := strconv.ParseUint(userIdString, 10, 64)
	if err != nil {
		fmt.Fprint(writer, err)
		return
	}

	currentUserId := context.Get(request, "userId").(uint64)
	if currentUserId != userId {
		currentUser, err := usersHandler.resourcesConnection.Users.Read.Get(currentUserId)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
			return
		} else if !currentUser.IsAdmin {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(writer, "Forbidden request for user %d", userId)
			return
		}
	}

	sshKeys, err := usersHandler.resourcesConnection.Users.Read.GetKeys(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedSshKeys := make([]sanitizedSshKey, 0, 10)
	for _, sshKey := range sshKeys {
		sanitizedSshKeys = append(sanitizedSshKeys, *getSanitizedSshKey(&sshKey))
	}

	jsonedSshKeys, err := json.Marshal(sanitizedSshKeys)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedSshKeys)
}
