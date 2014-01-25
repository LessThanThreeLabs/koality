package internalapi

import (
	"koality/resources"
)

type RepositoryReader struct {
	resourcesConnection *resources.Connection
	repositoriesPath    string
}

type RepositoryInfo struct {
       resources.Repository
	RepositoriesPath string
}

func (repositoryReader RepositoryReader) GetRepoFromLocalUri(localUri string, repositoryInfo *RepositoryInfo) error {
	repository, err := repositoryReader.resourcesConnection.Repositories.Read.GetByLocalUri(localUri)
	if err != nil {
		return err
	}

	repositoryInfo.Repository = *repository
	repositoryInfo.RepositoriesPath = repositoryReader.repositoriesPath
	return nil
}
