package google

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/resources"
)

type UserInformation struct {
	Id            string      `json:"id"`
	Email         string      `json:"email"`
	VerifiedEmail interface{} `json:"verified_email"` // Sometimes google returns a string, other times a boolean
	Name          string      `json:"name"`
	GivenName     string      `json:"given_name"`
	FamilyName    string      `json:"family_name"`
}

type GoogleHandler struct {
	resourcesConnection *resources.Connection
	sessionStore        sessions.Store
	sessionName         string
}

func New(resourcesConnection *resources.Connection, sessionStore sessions.Store, sessionName string) (*GoogleHandler, error) {
	return &GoogleHandler{resourcesConnection, sessionStore, sessionName}, nil
}

func (googleHandler *GoogleHandler) WireOAuthSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/token", googleHandler.handleOAuthToken).Methods("GET")
}
