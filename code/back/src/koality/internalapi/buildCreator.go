package internalapi

import (
	"koality/repositorymanager"
	"koality/resources"
)

type BuildCreator struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
}

type CreateBuildArg struct {
	RepositoryName string
	Ref            string
	BaseSha        string
	HeadSha        string
}

func (buildCreator BuildCreator) CreateBuild(createBuildArg CreateBuildArg, buildId *uint64) error {
	repository, err := buildCreator.resourcesConnection.Repositories.Read.GetByName(createBuildArg.RepositoryName)
	if err != nil {
		return err
	}

	_, headMessage, headUsername, headEmail, err := buildCreator.repositoryManager.GetCommitAttributes(repository, createBuildArg.HeadSha)
	if err != nil {
		return err
	}

	build, err := buildCreator.resourcesConnection.Builds.Create.Create(repository.Id, createBuildArg.HeadSha, createBuildArg.BaseSha, headMessage, headUsername, headEmail, nil, headEmail, createBuildArg.Ref, true, true)
	if err != nil {
		return err
	}

	*buildId = build.Id
	return nil
}
