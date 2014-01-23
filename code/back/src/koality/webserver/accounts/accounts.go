package accounts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/resources"
)

const (
	rememberMeDuration = 2592000
)

type AccountsHandler struct {
	resourcesConnection *resources.Connection
	sessionStore        sessions.Store
	sessionName         string
	passwordHasher      *resources.PasswordHasher
}

func New(resourcesConnection *resources.Connection, sessionStore sessions.Store, sessionName string, passwordHasher *resources.PasswordHasher) (*AccountsHandler, error) {
	return &AccountsHandler{resourcesConnection, sessionStore, sessionName, passwordHasher}, nil
}

func (accountsHandler *AccountsHandler) WireSubroutes(subrouter *mux.Router) {
	fmt.Println("...need to add functionality to reset password")
	subrouter.HandleFunc("/login", accountsHandler.Login).Methods("POST")
	subrouter.HandleFunc("/logout", accountsHandler.Logout).Methods("POST")
}
