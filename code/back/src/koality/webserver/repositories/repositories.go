package repositories

import (
	"github.com/gorilla/mux"
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
	LocalUri  string                             `json:"localUri"`
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
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager) (*RepositoriesHandler, error) {
	return &RepositoriesHandler{resourcesConnection, repositoryManager}, nil
}

func (repositoriesHandler *RepositoriesHandler) WireAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{repositoryId:[0-9]+}",
		middleware.IsLoggedInWrapper(repositoriesHandler.Get)).
		Methods("GET")
	subrouter.HandleFunc("/",
		middleware.IsLoggedInWrapper(repositoriesHandler.GetAll)).
		Methods("GET")

	subrouter.HandleFunc("/create",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.Create)).
		Methods("POST")

	subrouter.HandleFunc("/gitHub/create",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.CreateWithGitHub)).
		Methods("POST")

	subrouter.HandleFunc("/{repositoryId:[0-9]+}/gitHub/setHook",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.SetGitHubHookTypes)).
		Methods("PUT")
	subrouter.HandleFunc("/{repositoryId:[0-9]+}/gitHub/clearHook",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.ClearGitHubHook)).
		Methods("PUT")

	subrouter.HandleFunc("/{repositoryId:[0-9]+}",
		middleware.IsAdminWrapper(repositoriesHandler.resourcesConnection, repositoriesHandler.Delete)).
		Methods("DELETE")
}

func getSanitizedRepository(repository *resources.Repository) *sanitizedRepository {
	return &sanitizedRepository{
		Id:        repository.Id,
		Name:      repository.Name,
		Status:    repository.Status,
		VcsType:   repository.VcsType,
		LocalUri:  repository.LocalUri,
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
