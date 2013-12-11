package repositories

import (
	"database/sql"
	"koality/resources"
)

type CreateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewCreateHandler(database *sql.DB, verifier *Verifier) (resources.RepositoriesCreateHandler, error) {
	return &CreateHandler{database, verifier}, nil
}

func (createHandler *CreateHandler) Create(name, vcsType, localUri, remoteUri string) (uint64, error) {
	err := createHandler.getRepositoryParamsError(name, vcsType, localUri, remoteUri)
	if err != nil {
		return 0, err
	}

	id := uint64(0)
	query := "INSERT INTO repositories (name, vcs_type, local_uri, remote_uri) VALUES ($1, $2, $3, $4) RETURNING id"
	err = createHandler.database.QueryRow(query, name, vcsType, localUri, remoteUri).Scan(&id)
	return id, err
}

func (createHandler *CreateHandler) CreateWithGitHub(name, vcsType, localUri, remoteUri, gitHubOwner, gitHubName string) (uint64, error) {
	err := createHandler.getRepositoryParamsError(name, vcsType, localUri, remoteUri)
	if err != nil {
		return 0, err
	}

	transaction, err := createHandler.database.Begin()
	if err != nil {
		return 0, err
	}

	id := uint64(0)
	repositoryQuery := "INSERT INTO repositories (name, vcs_type, local_uri, remote_uri) VALUES ($1, $2, $3, $4) RETURNING id"
	err = transaction.QueryRow(repositoryQuery, name, vcsType, localUri, remoteUri).Scan(&id)
	if err != nil {
		transaction.Rollback()
		return 0, err
	}

	gitHubQuery := "INSERT INTO repository_github_metadatas (repository_id, owner, name) VALUES ($1, $2, $3)"
	_, err = transaction.Exec(gitHubQuery, id, gitHubOwner, gitHubName)
	if err != nil {
		transaction.Rollback()
		return 0, err
	}

	transaction.Commit()
	return id, nil
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
