package users

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"koality/repositorymanager"
	"koality/resources"
	"net/http"
	"strconv"
	"time"
)

type sanitizedUser struct {
	Id        uint64     `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	IsAdmin   bool       `json:"isAdmin"`
	Created   *time.Time `json:"created"`
}

type sanitizedSshKey struct {
	Name      string     `json:"name"`
	PublicKey string     `json:"publicKey"`
	Created   *time.Time `json:"created"`
}

type UsersHandler struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager) (*UsersHandler, error) {
	return &UsersHandler{resourcesConnection, repositoryManager}, nil
}

func (usersHandler *UsersHandler) WireSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{userId:[0-9]+}", usersHandler.Get).Methods("GET")
	subrouter.HandleFunc("/{userId:[0-9]+}/keys", usersHandler.GetKeys).Methods("GET")
	subrouter.HandleFunc("/all", usersHandler.GetAll).Methods("GET")
}

func getSanitizedUser(user *resources.User) *sanitizedUser {
	return &sanitizedUser{
		Id:        user.Id,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsAdmin:   user.IsAdmin,
		Created:   user.Created,
	}
}

func getSanitizedSshKey(sshKey *resources.SshKey) *sanitizedSshKey {
	return &sanitizedSshKey{
		Name:      sshKey.Name,
		PublicKey: sshKey.PublicKey,
		Created:   sshKey.Created,
	}
}

func (usersHandler *UsersHandler) hasAccessToUser(userId, requestedUserId uint64) (bool, error) {
	if userId == requestedUserId {
		return true, nil
	}

	user, err := usersHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		return false, err
	}
	return user.IsAdmin, nil
}

func (usersHandler *UsersHandler) Get(writer http.ResponseWriter, request *http.Request) {
	userIdString := mux.Vars(request)["userId"]
	userId, err := strconv.ParseUint(userIdString, 10, 64)
	if err != nil {
		fmt.Fprint(writer, err)
		return
	}

	id := context.Get(request, "userId").(uint64)
	allowedToMakeRequest, err := usersHandler.hasAccessToUser(id, userId)
	if err != nil {
		fmt.Fprint(writer, err)
		return
	} else if !allowedToMakeRequest {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Disallowed request")
		return
	}

	user, err := usersHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		fmt.Fprint(writer, err)
		return
	}

	sanitizedUser := getSanitizedUser(user)
	jsonedUser, err := json.MarshalIndent(sanitizedUser, "", "\t")
	if err != nil {
		fmt.Print(writer, err)
	}

	fmt.Fprintf(writer, "%s", jsonedUser)
}

func (usersHandler *UsersHandler) GetAll(writer http.ResponseWriter, request *http.Request) {
	users, err := usersHandler.resourcesConnection.Users.Read.GetAll()
	if err != nil {
		fmt.Fprint(writer, err)
		return
	}

	sanitizedUsers := make([]sanitizedUser, 0, 10)
	for _, user := range users {
		sanitizedUsers = append(sanitizedUsers, *getSanitizedUser(&user))
	}

	jsonedUsers, err := json.MarshalIndent(sanitizedUsers, "", "\t")
	if err != nil {
		fmt.Print(writer, err)
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

	sshKeys, err := usersHandler.resourcesConnection.Users.Read.GetKeys(userId)
	if err != nil {
		fmt.Fprint(writer, err)
		return
	}

	sanitizedSshKeys := make([]sanitizedSshKey, 0, 10)
	for _, sshKey := range sshKeys {
		sanitizedSshKeys = append(sanitizedSshKeys, *getSanitizedSshKey(&sshKey))
	}

	jsonedSshKeys, err := json.MarshalIndent(sanitizedSshKeys, "", "\t")
	if err != nil {
		fmt.Print(writer, err)
	}

	fmt.Fprintf(writer, "%s", jsonedSshKeys)
}
