package verifications

import (
	"github.com/gorilla/mux"
	"koality/repositorymanager"
	"koality/resources"
	"time"
)

type sanitizedVerification struct {
	Id           uint64     `json:"id"`
	RepositoryId uint64     `json:"verificationId"`
	SnapshotId   uint64     `json:"snapshotId"`
	MergeTarget  string     `json:"mergeTarget"`
	Status       string     `json:"status"`
	MergeStatus  string     `json:"mergeStatus"`
	Created      *time.Time `json:"created"`
	Started      *time.Time `json:"started"`
	Ended        *time.Time `json:"ended"`
	Changeset    sanitizedChangeset
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

func (verificationsHandler *VerificationsHandler) WireSubroutes(subrouter *mux.Router) {
	// subrouter.HandleFunc("/{verificationId:[0-9]+", verificationsHandler.Get).Methods("GET")
	// subrouter.HandleFunc("/tail", verificationsHandler.GetTail).Methods("GET")

	// subrouter.HandleFunc("/{verificationId:[0-9]+/retrigger",
	// 	middleware.IsAdminWrapper(verificationsHandler.resourcesConnection, verificationsHandler.Retrigger)).
	// 	Methods("POST")
}

func getSanitizedVerification(verification *resources.Verification) *sanitizedVerification {
	return &sanitizedVerification{
		Id:           verification.Id,
		RepositoryId: verification.RepositoryId,
		SnapshotId:   verification.SnapshotId,
		MergeTarget:  verification.MergeTarget,
		Status:       verification.Status,
		MergeStatus:  verification.MergeStatus,
		Created:      verification.Created,
		Started:      verification.Started,
		Ended:        verification.Ended,
		Changeset:    getSanitizedChangeset(verification.Changeset),
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
