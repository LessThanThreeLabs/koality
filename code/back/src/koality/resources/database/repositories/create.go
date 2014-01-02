package repositories

import (
	"database/sql"
	"koality/resources"
)

const (
	initialRepositoryStatus = "declared"
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	readHandler         resources.RepositoriesReadHandler
	subscriptionHandler resources.InternalRepositoriesSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.RepositoriesReadHandler,
	subscriptionHandler resources.InternalRepositoriesSubscriptionHandler) (resources.RepositoriesCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) Create(name, vcsType, localUri, remoteUri string) (*resources.Repository, error) {
	err := createHandler.getRepositoryParamsError(name, vcsType, localUri, remoteUri)
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	query := "INSERT INTO repositories (name, status, vcs_type, local_uri, remote_uri) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = createHandler.database.QueryRow(query, name, initialRepositoryStatus, vcsType, localUri, remoteUri).Scan(&id)
	if err != nil {
		return nil, err
	}

	repository, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(repository)
	return repository, nil
}

func (createHandler *CreateHandler) CreateWithGitHub(name, vcsType, localUri, remoteUri, gitHubOwner, gitHubName string) (*resources.Repository, error) {
	err := createHandler.getRepositoryParamsError(name, vcsType, localUri, remoteUri)
	if err != nil {
		return nil, err
	}

	transaction, err := createHandler.database.Begin()
	if err != nil {
		return nil, err
	}

	id := uint64(0)
	repositoryQuery := "INSERT INTO repositories (name, status, vcs_type, local_uri, remote_uri) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = transaction.QueryRow(repositoryQuery, name, initialRepositoryStatus, vcsType, localUri, remoteUri).Scan(&id)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	gitHubQuery := "INSERT INTO repository_github_metadatas (repository_id, owner, name) VALUES ($1, $2, $3)"
	_, err = transaction.Exec(gitHubQuery, id, gitHubOwner, gitHubName)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	transaction.Commit()

	repository, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(repository)
	return repository, nil
}

func (createHandler *CreateHandler) getRepositoryParamsError(name, vcsType, localUri, remoteUri string) error {
	if err := createHandler.verifier.verifyName(name); err != nil {
		return err
	}
	if err := createHandler.verifier.verifyVcsType(vcsType); err != nil {
		return err
	}
	if vcsType == "git" {
		if err := createHandler.verifier.verifyLocalGitUri(localUri); err != nil {
			return err
		}
		if err := createHandler.verifier.verifyRemoteGitUri(remoteUri); err != nil {
			return err
		}
	}
	if vcsType == "hg" {
		if err := createHandler.verifier.verifyLocalHgUri(localUri); err != nil {
			return err
		}
		if err := createHandler.verifier.verifyRemoteHgUri(remoteUri); err != nil {
			return err
		}
	}
	return nil
}
