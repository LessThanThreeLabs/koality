package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"regexp"
)

const (
	repositoryMaxNameLength      = 256
	repositoryMaxLocalUriLength  = 256
	repositoryMaxRemoteUriLength = 256
	nameRegex                    = "^[-_a-zA-Z0-9]+$"
	gitUriRegex                  = "[-_\\./a-zA-Z0-9]+@[-_\\.:/a-zA-Z0-9]+$"
	hgUriRegex                   = "[-_\\./a-zA-Z0-9]+@[-_\\./a-zA-Z0-9]+$"
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyName(name string) error {
	if len(name) > repositoryMaxNameLength {
		errorText := fmt.Sprintf("Name cannot exceed %d characters long", repositoryMaxNameLength)
		return errors.New(errorText)
	} else if ok, err := regexp.MatchString(nameRegex, name); !ok || err != nil {
		return errors.New("Name must match regex: " + nameRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithName(name); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyVcsType(vcsType string) error {
	if vcsType != "git" && vcsType != "hg" {
		return errors.New("Repository type must be 'git' or 'hg'")
	}
	return nil
}

func (verifier *Verifier) verifyLocalGitUri(localUri string) error {
	if len(localUri) > repositoryMaxLocalUriLength {
		errorText := fmt.Sprintf("Git local uri cannot exceed %d characters long", repositoryMaxLocalUriLength)
		return errors.New(errorText)
	} else if ok, err := regexp.MatchString(gitUriRegex, localUri); !ok || err != nil {
		return errors.New("Git local uri must match regex: " + gitUriRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithLocalUri(localUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyRemoteGitUri(remoteUri string) error {
	if len(remoteUri) > repositoryMaxRemoteUriLength {
		errorText := fmt.Sprintf("Git remote uri cannot exceed %d characters long", repositoryMaxRemoteUriLength)
		return errors.New(errorText)
	} else if ok, err := regexp.MatchString(gitUriRegex, remoteUri); !ok || err != nil {
		return errors.New("Git local uri must match regex: " + gitUriRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithRemoteUri(remoteUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyLocalHgUri(localUri string) error {
	if len(localUri) > repositoryMaxLocalUriLength {
		errorText := fmt.Sprintf("Hg local uri cannot exceed %d characters long", repositoryMaxLocalUriLength)
		return errors.New(errorText)
	} else if ok, err := regexp.MatchString(hgUriRegex, localUri); !ok || err != nil {
		return errors.New("Hg local uri must match regex: " + hgUriRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithLocalUri(localUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyRemoteHgUri(remoteUri string) error {
	if len(remoteUri) > repositoryMaxRemoteUriLength {
		errorText := fmt.Sprintf("Hg remote uri cannot exceed %d characters long", repositoryMaxRemoteUriLength)
		return errors.New(errorText)
	} else if ok, err := regexp.MatchString(hgUriRegex, remoteUri); !ok || err != nil {
		return errors.New("Hg local uri must match regex: " + hgUriRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithRemoteUri(remoteUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyRepositoryDoesNotExistWithName(name string) error {
	query := "SELECT id FROM repositories WHERE name=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, name).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := "Repository already exists with name: " + name
		return resources.RepositoryAlreadyExistsError{errorText}
	}
	return nil
}

func (verifier *Verifier) verifyRepositoryDoesNotExistWithLocalUri(localUri string) error {
	query := "SELECT id FROM repositories WHERE local_uri=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, localUri).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := "Repository already exists with local uri: " + localUri
		return resources.RepositoryAlreadyExistsError{errorText}
	}
	return nil
}

func (verifier *Verifier) verifyRepositoryDoesNotExistWithRemoteUri(remoteUri string) error {
	query := "SELECT id FROM repositories WHERE remote_uri=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, remoteUri).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := "Repository already exists with remote uri: " + remoteUri
		return resources.RepositoryAlreadyExistsError{errorText}
	}
	return nil
}
