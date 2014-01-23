package accounts

import (
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
	subrouter.HandleFunc("/login", accountsHandler.Login).Methods("POST")
}
