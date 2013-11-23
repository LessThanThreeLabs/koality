package repositories

import (
	"database/sql"
	"koality/resources"
)

type CreateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewCreateHandler(database *sql.DB) (resources.RepositoriesCreateHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}
	return &CreateHandler{database, verifier}, nil
}

func (createHandler *CreateHandler) Create(name, vcsType, localUri, remoteUri string) (int64, error) {
	err := createHandler.getRepositoryParamsError(name, vcsType, localUri, remoteUri)
	if err != nil {
		return -1, err
	}

	id := int64(0)
	query := "INSERT INTO repositories (name, vcs_type, local_uri, remote_uri) VALUES ($1, $2, $3, $4) RETURNING id"
	err = createHandler.database.QueryRow(query, name, vcsType, localUri, remoteUri).Scan(&id)
	if err != nil {
		return -1, err
	}
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
