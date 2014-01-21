package internalapi

import (
	"koality/resources"
)

type RepositoryReader struct {
	resourcesConnection *resources.Connection
}

func (repositoryReader RepositoryReader) GetRepoFromLocalUri(localUri string, repoRes *resources.Repository) error {
	repo, err := repositoryReader.resourcesConnection.Repositories.Read.GetByLocalUri(localUri)
	if err != nil {
		return err
	}

	*repoRes = *repo
	return nil
}
