package builds

import (
	"github.com/gorilla/mux"
	"koality/repositorymanager"
	"koality/resources"
	"koality/webserver/middleware"
	"time"
)

type sanitizedVerification struct {
	Id           uint64             `json:"id"`
	RepositoryId uint64             `json:"repositoryId"`
	MergeTarget  string             `json:"mergeTarget"`
	Status       string             `json:"status"`
	MergeStatus  string             `json:"mergeStatus"`
	Created      *time.Time         `json:"created"`
	Started      *time.Time         `json:"started"`
	Ended        *time.Time         `json:"ended"`
	Changeset    sanitizedChangeset `json:"changeset"`
}

type sanitizedChangeset struct {
	Id           uint64     `json:"id"`
	RepositoryId uint64     `json:"repositoryId"`
	HeadSha      string     `json:"headSha"`
	BaseSha      string     `json:"baseSha"`
	HeadMessage  string     `json:"headMessage"`
	HeadUsername string     `json:"headUsername"`
	HeadEmail    string     `json:"headEmail"`
	Created      *time.Time `json:"created"`
}

type VerificationsHandler struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager) (*VerificationsHandler, error) {
	return &VerificationsHandler{resourcesConnection, repositoryManager}, nil
}

func (buildsHandler *VerificationsHandler) WireAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{buildId:[0-9]+}",
		middleware.IsLoggedInWrapper(buildsHandler.get)).
		Methods("GET")
	subrouter.HandleFunc("/tail",
		middleware.IsLoggedInWrapper(buildsHandler.getTail)).
		Methods("GET")

	subrouter.HandleFunc("/{buildId:[0-9]+}/retrigger",
		middleware.IsLoggedInWrapper(buildsHandler.retrigger)).
		Methods("POST")
}

func (buildsHandler *VerificationsHandler) WireApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{buildId:[0-9]+}", buildsHandler.get).Methods("GET")
	subrouter.HandleFunc("/tail", buildsHandler.getTail).Methods("GET")

	subrouter.HandleFunc("/", buildsHandler.create).Methods("POST")
	subrouter.HandleFunc("/{buildId:[0-9]+}/retrigger", buildsHandler.retrigger).Methods("POST")
}

func getSanitizedVerification(build *resources.Verification) *sanitizedVerification {
	return &sanitizedVerification{
		Id:           build.Id,
		RepositoryId: build.RepositoryId,
		MergeTarget:  build.MergeTarget,
		Status:       build.Status,
		MergeStatus:  build.MergeStatus,
		Created:      build.Created,
		Started:      build.Started,
		Ended:        build.Ended,
		Changeset:    getSanitizedChangeset(build.Changeset),
	}
}

func getSanitizedChangeset(changeset resources.Changeset) sanitizedChangeset {
	return sanitizedChangeset{
		Id:           changeset.Id,
		RepositoryId: changeset.RepositoryId,
		HeadSha:      changeset.HeadSha,
		BaseSha:      changeset.BaseSha,
		HeadMessage:  changeset.HeadMessage,
		HeadUsername: changeset.HeadUsername,
		HeadEmail:    changeset.HeadEmail,
		Created:      changeset.Created,
	}
}
