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
	filePathRegex                = "^[-__\\./a-zA-Z0-9]+$"
	remoteGitUriRegex            = "[-_\\./a-zA-Z0-9]+@[-_\\.:/a-zA-Z0-9]+$"
	remoteHgUriRegex             = "[-_\\./a-zA-Z0-9]+@[-_\\./a-zA-Z0-9]+$"
)

var (
	allowedStatuses  []string = []string{"preparing", "installed"}
	allowedHookTypes []string = []string{"push", "pull_request"}
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyName(name string) error {
	if len(name) > repositoryMaxNameLength {
		return fmt.Errorf("Name cannot exceed %d characters long", repositoryMaxNameLength)
	} else if ok, err := regexp.MatchString(nameRegex, name); !ok || err != nil {
		return errors.New("Name must match regex: " + nameRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithName(name); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyStatus(status string) error {
	for _, allowedVerificationStatus := range allowedStatuses {
		if status == allowedVerificationStatus {
			return nil
		}
	}
	return resources.InvalidRepositoryStatusError{"Unexpected repository status: " + status}
}

func (verifier *Verifier) verifyVcsType(vcsType string) error {
	if vcsType != "git" && vcsType != "hg" {
		return errors.New("Repository type must be 'git' or 'hg'")
	}
	return nil
}

func (verifier *Verifier) verifyLocalGitUri(localUri string) error {
	if len(localUri) > repositoryMaxLocalUriLength {
		return fmt.Errorf("Git local uri cannot exceed %d characters long", repositoryMaxLocalUriLength)
	} else if ok, err := regexp.MatchString(filePathRegex, localUri); !ok || err != nil {
		return errors.New("Git local uri must match regex: " + filePathRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithLocalUri(localUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyRemoteGitUri(remoteUri string) error {
	if len(remoteUri) > repositoryMaxRemoteUriLength {
		return fmt.Errorf("Git remote uri cannot exceed %d characters long", repositoryMaxRemoteUriLength)
	} else if ok, err := regexp.MatchString(remoteGitUriRegex, remoteUri); !ok || err != nil {
		return errors.New("Git remote uri must match regex: " + remoteGitUriRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithRemoteUri(remoteUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyLocalHgUri(localUri string) error {
	if len(localUri) > repositoryMaxLocalUriLength {
		return fmt.Errorf("Hg local uri cannot exceed %d characters long", repositoryMaxLocalUriLength)
	} else if ok, err := regexp.MatchString(filePathRegex, localUri); !ok || err != nil {
		return errors.New("Hg local uri must match regex: " + filePathRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithLocalUri(localUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyRemoteHgUri(remoteUri string) error {
	if len(remoteUri) > repositoryMaxRemoteUriLength {
		return fmt.Errorf("Hg remote uri cannot exceed %d characters long", repositoryMaxRemoteUriLength)
	} else if ok, err := regexp.MatchString(remoteHgUriRegex, remoteUri); !ok || err != nil {
		return errors.New("Hg remote uri must match regex: " + remoteHgUriRegex)
	} else if err := verifier.verifyRepositoryDoesNotExistWithRemoteUri(remoteUri); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyHookTypes(hookTypes []string) error {
	hookTypeAllowed := func(hookType string) bool {
		for _, allowedHookType := range allowedHookTypes {
			if hookType == allowedHookType {
				return true
			}
		}
		return false
	}

	for _, hookType := range hookTypes {
		if !hookTypeAllowed(hookType) {
			return resources.InvalidRepositoryHookTypeError{"Unexpected hook type: " + hookType}
		}
	}
	return nil
}

func (verifier *Verifier) verifyRepositoryDoesNotExistWithName(name string) error {
	query := "SELECT id FROM repositories WHERE name=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, name).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		return resources.RepositoryAlreadyExistsError{"Repository already exists with name: " + name}
	}
	return nil
}

func (verifier *Verifier) verifyRepositoryDoesNotExistWithLocalUri(localUri string) error {
	query := "SELECT id FROM repositories WHERE local_uri=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, localUri).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		return resources.RepositoryAlreadyExistsError{"Repository already exists with local uri: " + localUri}
	}
	return nil
}

func (verifier *Verifier) verifyRepositoryDoesNotExistWithRemoteUri(remoteUri string) error {
	query := "SELECT id FROM repositories WHERE remote_uri=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, remoteUri).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		return resources.RepositoryAlreadyExistsError{"Repository already exists with remote uri: " + remoteUri}
	}
	return nil
}
