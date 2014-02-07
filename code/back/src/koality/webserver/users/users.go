package users

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"koality/webserver/middleware"
	"time"
)

type sanitizedUser struct {
	Id        uint64     `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	IsAdmin   bool       `json:"isAdmin"`
	Created   *time.Time `json:"created,omitempty"`
	IsDeleted bool       `json:"isDeleted"`
}

type sanitizedSshKey struct {
	Id        uint64     `json:"id"`
	Name      string     `json:"name"`
	PublicKey string     `json:"publicKey"`
	Created   *time.Time `json:"created,omitempty"`
}

type setNameRequestData struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type setPasswordRequestData struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type addKeyRequestData struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
}

type removeKeyRequestData struct {
	Id uint64 `json:"id"`
}

type UsersHandler struct {
	resourcesConnection *resources.Connection
	passwordHasher      *resources.PasswordHasher
}

func New(resourcesConnection *resources.Connection, passwordHasher *resources.PasswordHasher) (*UsersHandler, error) {
	return &UsersHandler{resourcesConnection, passwordHasher}, nil
}

func (usersHandler *UsersHandler) WireAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{userId:[0-9]+}",
		middleware.IsLoggedInWrapper(usersHandler.get)).
		Methods("GET")
	subrouter.HandleFunc("/",
		middleware.IsLoggedInWrapper(usersHandler.getAll)).
		Methods("GET")
	subrouter.HandleFunc("/keys",
		middleware.IsLoggedInWrapper(usersHandler.getKeys)).
		Methods("GET")

	subrouter.HandleFunc("/name",
		middleware.IsLoggedInWrapper(usersHandler.setName)).
		Methods("POST")
	subrouter.HandleFunc("/password",
		middleware.IsLoggedInWrapper(usersHandler.setPassword)).
		Methods("POST")
	subrouter.HandleFunc("/addKey",
		middleware.IsLoggedInWrapper(usersHandler.addKey)).
		Methods("POST")
	subrouter.HandleFunc("/removeKey",
		middleware.IsLoggedInWrapper(usersHandler.removeKey)).
		Methods("POST")

	subrouter.HandleFunc("/{userId:[0-9]+}/admin",
		middleware.IsAdminWrapper(usersHandler.resourcesConnection, usersHandler.setAdmin)).
		Methods("PUT")

	subrouter.HandleFunc("/{userId:[0-9]+}",
		middleware.IsAdminWrapper(usersHandler.resourcesConnection, usersHandler.delete)).
		Methods("DELETE")
}

func (usersHandler *UsersHandler) WireApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{userId:[0-9]+}", usersHandler.get).Methods("GET")
	subrouter.HandleFunc("/", usersHandler.getAll).Methods("GET")
	subrouter.HandleFunc("/{userId:[0-9]+}/keys", usersHandler.getKeys).Methods("GET")
}

func getSanitizedUser(user *resources.User) *sanitizedUser {
	return &sanitizedUser{
		Id:        user.Id,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsAdmin:   user.IsAdmin,
		Created:   user.Created,
		IsDeleted: user.IsDeleted,
	}
}

func getSanitizedSshKey(sshKey *resources.SshKey) *sanitizedSshKey {
	return &sanitizedSshKey{
		Id:        sshKey.Id,
		Name:      sshKey.Name,
		PublicKey: sshKey.PublicKey,
		Created:   sshKey.Created,
	}
}
