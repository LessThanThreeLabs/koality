package repositories

import (
	"database/sql"
	"errors"
	"koality/resources"
	"regexp"
)

const (
	nameRegex   = "^[-_a-zA-Z0-9]+$"
	gitUriRegex = "[-_\\./a-zA-Z0-9]+@[-_\\.:/a-zA-Z0-9]+$"
	hgUriRegex  = "[-_\\./a-zA-Z0-9]+@[-_\\./a-zA-Z0-9]+$"
)

type CreateHandler struct {
	database *sql.DB
}

func NewCreateHandler(database *sql.DB) (resources.RepositoriesCreateHandler, error) {
	return &CreateHandler{database}, nil
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
	if !createHandler.isValidName(name) {
		return errors.New("Name must be less than 256 characters and match: " + nameRegex)
	} else if !createHandler.isValidVcsType(vcsType) {
		return errors.New("Repository type must be git or hg")
	} else if vcsType == "git" && !createHandler.isValidGitUri(localUri) {
		return errors.New("Local uri must be less than 1024 characters and match: " + gitUriRegex)
	} else if vcsType == "git" && !createHandler.isValidGitUri(remoteUri) {
		return errors.New("Remote uri must be less than 1024 characters and match: " + gitUriRegex)
	} else if vcsType == "hg" && !createHandler.isValidHgUri(localUri) {
		return errors.New("Local uri must be less than 1024 characters and match: " + hgUriRegex)
	} else if vcsType == "hg" && !createHandler.isValidHgUri(remoteUri) {
		return errors.New("Remote uri must be less than 1024 characters and match: " + hgUriRegex)
	} else if createHandler.doesRepositoryExistWithName(name) {
		return resources.RepositoryAlreadyExistsError(errors.New("Repository already exists with name: " + name))
	} else if createHandler.doesRepositoryExistWithLocalUri(localUri) {
		return resources.RepositoryAlreadyExistsError(errors.New("Repository already exists with local uri: " + localUri))
	} else if createHandler.doesRepositoryExistWithRemoteUri(remoteUri) {
		return resources.RepositoryAlreadyExistsError(errors.New("Repository already exists with remote uri: " + remoteUri))
	}
	return nil
}

func (createHandler *CreateHandler) isValidName(name string) bool {
	if len(name) > 256 {
		return false
	} else if ok, err := regexp.MatchString(nameRegex, name); !ok || err != nil {
		return false
	}
	return true
}

func (createHandler *CreateHandler) isValidVcsType(vcsType string) bool {
	return vcsType == "git" || vcsType == "hg"
}

func (createHandler *CreateHandler) isValidGitUri(uri string) bool {
	if len(uri) > 256 {
		return false
	} else if ok, err := regexp.MatchString(gitUriRegex, uri); !ok || err != nil {
		return false
	}
	return true
}

func (createHandler *CreateHandler) isValidHgUri(uri string) bool {
	if len(uri) > 256 {
		return false
	} else if ok, err := regexp.MatchString(hgUriRegex, uri); !ok || err != nil {
		return false
	}
	return true
}

func (createHandler *CreateHandler) doesRepositoryExistWithName(name string) bool {
	query := "SELECT id FROM repositories WHERE name=$1 AND deleted=0"
	err := createHandler.database.QueryRow(query, name).Scan()
	return err != sql.ErrNoRows
}

func (createHandler *CreateHandler) doesRepositoryExistWithLocalUri(localUri string) bool {
	query := "SELECT id FROM repositories WHERE local_uri=$1 AND deleted=0"
	err := createHandler.database.QueryRow(query, localUri).Scan()
	return err != sql.ErrNoRows
}

func (createHandler *CreateHandler) doesRepositoryExistWithRemoteUri(remoteUri string) bool {
	query := "SELECT id FROM repositories WHERE remote_uri=$1 AND deleted=0"
	err := createHandler.database.QueryRow(query, remoteUri).Scan()
	return err != sql.ErrNoRows
}
