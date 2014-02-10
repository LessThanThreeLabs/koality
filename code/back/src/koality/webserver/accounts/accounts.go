package accounts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/mail"
	"koality/resources"
	"koality/webserver/middleware"
)

const (
	rememberMeDuration = 2592000
)

type loginRequestData struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

type AccountsHandler struct {
	resourcesConnection *resources.Connection
	sessionStore        sessions.Store
	sessionName         string
	passwordHasher      *resources.PasswordHasher
	mailer              mail.Mailer
}

func New(resourcesConnection *resources.Connection, sessionStore sessions.Store, sessionName string, passwordHasher *resources.PasswordHasher, mailer mail.Mailer) (*AccountsHandler, error) {
	return &AccountsHandler{resourcesConnection, sessionStore, sessionName, passwordHasher, mailer}, nil
}

func (accountsHandler *AccountsHandler) WireAppSubroutes(subrouter *mux.Router) {
	fmt.Println("...need to add functionality to reset password")

	subrouter.HandleFunc("/googleLoginRedirect",
		middleware.IsLoggedOutWrapper(accountsHandler.getGoogleLoginRedirect)).
		Methods("GET")
	subrouter.HandleFunc("/googleCreateAccountRedirect",
		middleware.IsLoggedOutWrapper(accountsHandler.getGoogleCreateAccountRedirect)).
		Methods("GET")

	subrouter.HandleFunc("/create",
		middleware.IsLoggedOutWrapper(accountsHandler.create)).
		Methods("POST")
	subrouter.HandleFunc("/gitHub/Create",
		middleware.IsLoggedOutWrapper(accountsHandler.createWithGitHub)).
		Methods("POST")
	subrouter.HandleFunc("/login",
		middleware.IsLoggedOutWrapper(accountsHandler.login)).
		Methods("POST")
	subrouter.HandleFunc("/logout",
		middleware.IsLoggedInWrapper(accountsHandler.logout)).
		Methods("POST")
}
