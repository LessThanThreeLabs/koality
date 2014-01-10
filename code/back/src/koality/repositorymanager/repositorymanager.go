package repositorymanager

import (
	"fmt"
	"koality/repositorymanager/repositorystore"
	"koality/resources"
)

func openPostPushRepository(repository *resources.Repository) (repositorystore.PostPushRepository, error) {
	switch repository.VcsType {
	case "git":
		return repositorystore.OpenGitRepository(repository), nil
	case "hg":
		return repositorystore.OpenHgRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}

}

func openPrePushRepository(repository *resources.Repository) (repositorystore.PrePushRepository, error) {
	switch repository.VcsType {
	case "git":
		return repositorystore.OpenGitRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}
}

func openStoredRepository(repository *resources.Repository) (repositorystore.StoredRepository, error) {
	switch repository.VcsType {
	case "git":
		return repositorystore.OpenGitRepository(repository), nil
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

func CreateRepository(repository *resources.Repository) (err error) {
	openedRepository, err := openStoredRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.CreateRepository()
}

func DeleteRepository(repository *resources.Repository) (err error) {
	openedRepository, err := openStoredRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.DeleteRepository()
}

func StorePending(repository *resources.Repository, ref string, args ...string) (err error) {
	openedRepository, err := openPrePushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.StorePending(ref, repository.RemoteUri, args...)
}

func MergeChangeset(repository *resources.Repository, headRef, baseRef, refToMergeInto string) (err error) {
	openedRepository, err := openPrePushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.MergeChangeset(headRef, baseRef, refToMergeInto)
}
