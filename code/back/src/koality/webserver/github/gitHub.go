package github

import (
	"github.com/gorilla/mux"
	"koality/repositorymanager"
	"koality/resources"
)

type PullRequestHookPayload struct {
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
}

type PullRequest struct {
	Head GitHubRef `json:"head"`
	Base GitHubRef `json:"base"`
}

type GitHubRef struct {
	Label      string                `json:"label"`
	Ref        string                `json:"ref"`
	Sha        string                `json:"sha"`
	User       PullRequestUser       `json:"user"`
	Repository PullRequestRepository `json:"repo"`
}

type PullRequestUser struct {
	Login string `json:"login"`
}

type PullRequestRepository struct {
	Name  string          `json:"name"`
	Owner PullRequestUser `json:"owner"`
}

type PushHookPayload struct {
	Before     string         `json:"before"`
	After      string         `json:"after"`
	Ref        string         `json:"ref"`
	Repository PushRepository `json:"repository"`
}

type PushRepository struct {
	Name  string   `json:"name"`
	Owner PushUser `json:"owner"`
}

type PushUser struct {
	Name string `json:"name"`
}

type GitHubHandler struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager) (*GitHubHandler, error) {
	return &GitHubHandler{resourcesConnection, repositoryManager}, nil
}

func (gitHubHandler *GitHubHandler) WireHooksSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/verifyChange", gitHubHandler.verifyChange).Methods("POST")
}

func (gitHubHandler *GitHubHandler) WireOAuthSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/token", gitHubHandler.handleOAuthToken).Methods("GET")
}
