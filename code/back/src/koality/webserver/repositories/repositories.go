package repositories

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"time"
)

type sanitizedRepository struct {
	Id        uint64 `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	VcsType   string `json:"vcsType"`
	LocalUri  string `json:"localUri"`
	RemoteUri string `json:"remoteUri"`
	GitHub    *sanitizedRepositoryGitHubMetadata
	Created   *time.Time `json:"created"`
	IsDeleted bool       `json:"isDeleted"`
}

type sanitizedRepositoryGitHubMetadata struct {
	Owner     string   `json:"owner"`
	Name      string   `json:"name"`
	HookTypes []string `json:"hookTypes"`
}

type RepositoriesHandler struct {
	resourcesConnection *resources.Connection
}

func New(resourcesConnection *resources.Connection) (*RepositoriesHandler, error) {
	return &RepositoriesHandler{resourcesConnection}, nil
}

func (repositoriesHandler *RepositoriesHandler) WireSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{repositoryId:[0-9]+}", repositoriesHandler.Get).Methods("GET")
	subrouter.HandleFunc("/", repositoriesHandler.GetAll).Methods("GET")

	// subrouter.HandleFunc("/name", usersHandler.SetName).Methods("POST")
	// subrouter.HandleFunc("/password", usersHandler.SetPassword).Methods("POST")
	// subrouter.HandleFunc("/addKey", usersHandler.AddKey).Methods("POST")
	// subrouter.HandleFunc("/removeKey", usersHandler.RemoveKey).Methods("POST")

	// subrouter.HandleFunc("/{repositoryId:[0-9]+}/admin", usersHandler.SetAdmin).Methods("PUT")

	// subrouter.HandleFunc("/{repositoryId:[0-9]+}", usersHandler.Delete).Methods("DELETE")
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
