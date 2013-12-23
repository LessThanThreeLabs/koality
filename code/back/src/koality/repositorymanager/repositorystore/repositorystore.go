package repositorystore

import (
	"fmt"
	"koality/repositorymanager/pathgenerator"
	"koality/resources"
	"os"
)

func checkRepositoryExists(repository *resources.Repository) (path string, err error) {
	path = pathgenerator.ToPath(repository)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return "", NoSuchRepositoryInStoreError{fmt.Sprintf("The repository %v could not be found in the repository store.", repository.Name)}
	}

	return
}

func CreateRepository(repository *resources.Repository) (err error) {
	vcsDispatcher := map[string]func(*resources.Repository) error{
		"git": gitCreateRepository,
		"hg":  hgCreateRepository,
	}

	return vcsDispatcher[repository.VcsType](repository)
}

func DeleteRepository(repository *resources.Repository) (err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return err
	}

	return os.RemoveAll(path)
}

func RenameRepository(repository *resources.Repository, newName string) (err error) {
	path, err := checkRepositoryExists(repository)
	if err != nil {
		return err
	}

	return os.Rename(path, newName)
}

func MergeChangeset(repository *resources.Repository, headRef, baseRef, mergeIntoRef string) error {
	return gitMergeChangeset(repository, headRef, baseRef, mergeIntoRef)
}

func GetCommitAttributes(repository *resources.Repository, ref string) (string, string, string, error) {
	vcsDispatcher := map[string]func(*resources.Repository, string) (string, string, string, error){
		"git": gitGetCommitAttributes,
		"hg":  hgGetCommitAttributes,
	}

	return vcsDispatcher[repository.VcsType](repository, ref)
}

func GetYamlFile(repository *resources.Repository, ref string) (string, error) {
	vcsDispatcher := map[string]func(*resources.Repository, string) (string, error){
		"git": gitGetYamlFile,
		"hg":  hgGetYamlFile,
	}

	return vcsDispatcher[repository.VcsType](repository, ref)
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

func (err RepositoryAlreadyExistsInStoreError) Error() string {
	return err.Message
}

func (err NoSuchRepositoryInStoreError) Error() string {
	return err.Message
}

func (err BadRepositorySetupError) Error() string {
	return err.Message
}
