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
		fmt.Fprintf(writer, "Unable to parse userId: %v", err)
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
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedUser)
}

func (usersHandler *UsersHandler) GetAll(writer http.ResponseWriter, request *http.Request) {
	users, err := usersHandler.resourcesConnection.Users.Read.GetAll()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedUsers := make([]sanitizedUser, 0, len(users))
	for _, user := range users {
		sanitizedUsers = append(sanitizedUsers, *getSanitizedUser(&user))
	}

	jsonedUsers, err := json.Marshal(sanitizedUsers)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedUsers)
}

func (usersHandler *UsersHandler) GetKeys(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	sshKeys, err := usersHandler.resourcesConnection.Users.Read.GetKeys(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedSshKeys := make([]sanitizedSshKey, 0, len(sshKeys))
	for _, sshKey := range sshKeys {
		sanitizedSshKeys = append(sanitizedSshKeys, *getSanitizedSshKey(&sshKey))
	}

	jsonedSshKeys, err := json.Marshal(sanitizedSshKeys)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedSshKeys)
}
