package accounts

import (
	"encoding/gob"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/mail"
	"koality/resources"
	"koality/webserver/middleware"
)

const (
	confirmAccountFlashKey = "createAccountData"
)

type createAccountRequestData struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

type confirmAccountData struct {
	Email        string
	FirstName    string
	LastName     string
	PasswordHash []byte
	PasswordSalt []byte
	Token        string
}

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

func init() {
	gob.Register(&confirmAccountData{})
}

func New(resourcesConnection *resources.Connection, sessionStore sessions.Store, sessionName string, passwordHasher *resources.PasswordHasher, mailer mail.Mailer) (*AccountsHandler, error) {
	return &AccountsHandler{resourcesConnection, sessionStore, sessionName, passwordHasher, mailer}, nil
}

func (accountsHandler *AccountsHandler) WireAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/googleLoginRedirect",
		middleware.IsLoggedOutWrapper(accountsHandler.getGoogleLoginRedirect)).
		Methods("GET")
	subrouter.HandleFunc("/googleCreateAccountRedirect",
		middleware.IsLoggedOutWrapper(accountsHandler.getGoogleCreateAccountRedirect)).
		Methods("GET")

	subrouter.HandleFunc("/create",
		middleware.IsLoggedOutWrapper(accountsHandler.create)).
		Methods("POST")
	subrouter.HandleFunc("/confirm",
		middleware.IsLoggedOutWrapper(accountsHandler.confirm)).
		Methods("POST")
	subrouter.HandleFunc("/login",
		middleware.IsLoggedOutWrapper(accountsHandler.login)).
		Methods("POST")
	subrouter.HandleFunc("/logout",
		middleware.IsLoggedInWrapper(accountsHandler.logout)).
		Methods("POST")
	subrouter.HandleFunc("/resetPassword",
		middleware.IsLoggedOutWrapper(accountsHandler.resetPassword)).
		Methods("POST")
}
