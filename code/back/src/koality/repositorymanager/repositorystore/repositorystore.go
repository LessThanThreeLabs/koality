package repositorystore

import (
	"fmt"
	"os"
)

const (
	//TODO(akostov or bbland) hook these up.
	defaultTimeout        = 120
	defaultSshScript      = ""
	defaultPrivateKeyPath = ""
)

func checkRepositoryExists(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return NoSuchRepositoryInStoreError{fmt.Sprintf("There os no repository at %s could not be found in the repository store.", path)}
	}

	return
}

type RepositoryAlreadyExistsInStoreError struct {
	Message string
}

type NoSuchRepositoryInStoreError struct {
	Message string
}

type BadRepositorySetupError struct {
	Message string
}

type NoSuchCommitInRepositoryError struct {
	Message string
}

func (err RepositoryAlreadyExistsInStoreError) Error() string {
	return err.Message
}

func (err NoSuchRepositoryInStoreError) Error() string {
	return err.Message
}

func (err BadRepositorySetupError) Error() string {
	return err.Message
}

func (err NoSuchCommitInRepositoryError) Error() string {
	return err.Message
}
