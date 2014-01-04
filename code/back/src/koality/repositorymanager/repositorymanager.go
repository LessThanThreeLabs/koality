package repositorymanager

import (
	"fmt"
	"koality/repositorymanager/repositorystore"
	"koality/resources"
)

type StoredRepository interface {
	CreateRepository() error

	DeleteRepository() error
}

type PostPushRepository interface {
	GetYamlFile(ref string) (string, error)

	GetCommitAttributes(ref string) (string, string, string, error)
}

type PrePushRepository interface {
	StoredRepository
	PostPushRepository

	StorePending(string, string, ...string) error
}

func openPostPushRepository(repository *resources.Repository) (PostPushRepository, error) {
	switch repository.VcsType {
	case "hg":
		return repositorystore.OpenHgRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}

}

func openPrePushRepository(repository *resources.Repository) (PrePushRepository, error) {
	switch repository.VcsType {
	case "git":
		return repositorystore.OpenGitRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}
}

func openStoredRepository(repository *resources.Repository) (StoredRepository, error) {
	switch repository.VcsType {
	case "hg":
		return repositorystore.OpenHgRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}
}

func GetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	openedRepository, err := openPostPushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.GetYamlFile(ref)
}

func GetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	openedRepository, err := openPostPushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.GetCommitAttributes(ref)
}
