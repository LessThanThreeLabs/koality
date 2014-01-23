package users

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"time"
)

type sanitizedUser struct {
	Id        uint64     `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	IsAdmin   bool       `json:"isAdmin"`
	Created   *time.Time `json:"created"`
	IsDeleted bool       `json:"isDeleted"`
}

type sanitizedSshKey struct {
	Name      string     `json:"name"`
	PublicKey string     `json:"publicKey"`
	Created   *time.Time `json:"created"`
}

type UsersHandler struct {
	resourcesConnection *resources.Connection
	passwordHasher      *resources.PasswordHasher
}

func New(resourcesConnection *resources.Connection, passwordHasher *resources.PasswordHasher) (*UsersHandler, error) {
	return &UsersHandler{resourcesConnection, passwordHasher}, nil
}

func (usersHandler *UsersHandler) WireSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{userId:[0-9]+}", usersHandler.Get).Methods("GET")
	subrouter.HandleFunc("/{userId:[0-9]+}/keys", usersHandler.GetKeys).Methods("GET")
	subrouter.HandleFunc("/", usersHandler.GetAll).Methods("GET")

	subrouter.HandleFunc("/name", usersHandler.SetName).Methods("POST")
	subrouter.HandleFunc("/password", usersHandler.SetPassword).Methods("POST")
	subrouter.HandleFunc("/addKey", usersHandler.AddKey).Methods("POST")
	subrouter.HandleFunc("/removeKey", usersHandler.RemoveKey).Methods("POST")

	subrouter.HandleFunc("/{userId:[0-9]+}/admin", usersHandler.SetAdmin).Methods("PUT")

	subrouter.HandleFunc("/{userId:[0-9]+}", usersHandler.Delete).Methods("DELETE")
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
		Name:      sshKey.Name,
		PublicKey: sshKey.PublicKey,
		Created:   sshKey.Created,
	}
}
