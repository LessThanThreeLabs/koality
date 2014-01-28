package accounts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/resources"
	"koality/webserver/middleware"
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

func (accountsHandler *AccountsHandler) WireAppSubroutes(subrouter *mux.Router) {
	fmt.Println("...need to add functionality to reset password")
	subrouter.HandleFunc("/login",
		middleware.IsLoggedOutWrapper(accountsHandler.Logout)).
		Methods("POST")
	subrouter.HandleFunc("/logout",
		middleware.IsLoggedInWrapper(accountsHandler.Logout)).
		Methods("POST")
}
