package repositories

import (
	"github.com/gorilla/mux"
	"koality/github"
	"koality/repositorymanager"
	"koality/resources"
	"koality/webserver/middleware"
	"time"
)

type sanitizedRepository struct {
	Id        uint64                             `json:"id"`
	Name      string                             `json:"name"`
	Status    string                             `json:"status"`
	VcsType   string                             `json:"vcsType"`
	RemoteUri string                             `json:"remoteUri"`
	GitHub    *sanitizedRepositoryGitHubMetadata `json:"gitHub,omitempty"`
	Created   *time.Time                         `json:"created"`
	IsDeleted bool                               `json:"isDeleted"`
}

type sanitizedRepositoryGitHubMetadata struct {
	Owner     string   `json:"owner"`
	Name      string   `json:"name"`
	HookTypes []string `json:"hookTypes"`
}

type RepositoriesHandler struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
	gitHubConnection    github.GitHubConnection
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager, gitHubConnection github.GitHubConnection) (*RepositoriesHandler, error) {
	return &RepositoriesHandler{resourcesConnection, repositoryManager, gitHubConnection}, nil
}

func (repositoriesHandler *RepositoriesHandler) WireAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{repositoryId:[0-9]+}",
		middleware.IsLoggedInWrapper(repositoriesHandler.get)).
		Methods("GET")
	subrouter.HandleFunc("/",
		middleware.IsLoggedInWrapper(repositoriesHandler.getAll)).
		Methods("GET")

	subrouter.HandleFunc("/create",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.create)).
		Methods("POST")
	subrouter.HandleFunc("/gitHub/create",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.createWithGitHub)).
		Methods("POST")

	subrouter.HandleFunc("/{repositoryId:[0-9]+}/gitHub/setHook",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.setGitHubHookTypes)).
		Methods("PUT")
	subrouter.HandleFunc("/{repositoryId:[0-9]+}/gitHub/clearHook",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.clearGitHubHook)).
		Methods("PUT")

	subrouter.HandleFunc("/{repositoryId:[0-9]+}",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.delete)).
		Methods("DELETE")
}

func (repositoriesHandler *RepositoriesHandler) WireApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{repositoryId:[0-9]+}", repositoriesHandler.get).Methods("GET")
	subrouter.HandleFunc("/", repositoriesHandler.getAll).Methods("GET")
}

func getSanitizedRepository(repository *resources.Repository) *sanitizedRepository {
	return &sanitizedRepository{
		Id:        repository.Id,
		Name:      repository.Name,
		Status:    repository.Status,
		VcsType:   repository.VcsType,
		RemoteUri: repository.RemoteUri,
		GitHub:    getSanitizedRepositoryGitHubMetadata(repository.GitHub),
		Created:   repository.Created,
		IsDeleted: repository.IsDeleted,
	}
}

func getSanitizedRepositoryGitHubMetadata(repositoryGitHubMetadata *resources.RepositoryGitHubMetadata) *sanitizedRepositoryGitHubMetadata {
	if repositoryGitHubMetadata == nil {
		return nil
	}

	return &sanitizedRepositoryGitHubMetadata{
		Owner:     repositoryGitHubMetadata.Owner,
		Name:      repositoryGitHubMetadata.Name,
		HookTypes: repositoryGitHubMetadata.HookTypes,
	}
}
