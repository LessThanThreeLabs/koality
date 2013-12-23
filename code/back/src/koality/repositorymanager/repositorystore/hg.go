package repositorystore

import (
	"fmt"
	"koality/repositorymanager/pathgenerator"
	"koality/resources"
	"os"
)

func hgCreateRepository(repository *resources.Repository) (err error) {
	path := pathgenerator.ToPath(repository)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return RepositoryAlreadyExistsInStoreError{fmt.Sprintf("The repository %v already exists in the repository store.", repository.Name)}
	}

	return
}

func hgGetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	return
}

func hgGetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	return
}
