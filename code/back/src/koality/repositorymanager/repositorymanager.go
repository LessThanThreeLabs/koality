package repositorymanager

import (
	"fmt"
	"koality/resources"
	"sync"
)

var lockMap = make(map[uint64]sync.Locker)
var lockMapMutex sync.Mutex

func openPostPushRepository(repository *resources.Repository) (PostPushRepository, error) {
	switch repository.VcsType {
	case "git":
		return openGitRepository(repository), nil
	case "hg":
		return openHgRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}

}

func openPrePushRepository(repository *resources.Repository) (PrePushRepository, error) {
	switch repository.VcsType {
	case "git":
		return openGitRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}
}

func openStoredRepository(repository *resources.Repository) (StoredRepository, error) {
	switch repository.VcsType {
	case "git":
		return openGitRepository(repository), nil
	case "hg":
		return openHgRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}
}

func GetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	openedRepository, err := openPostPushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.getYamlFile(ref)
}

func GetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	openedRepository, err := openPostPushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.getCommitAttributes(ref)
}

func CreateRepository(repository *resources.Repository) (err error) {
	openedRepository, err := openStoredRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.createRepository()
}

func DeleteRepository(repository *resources.Repository) (err error) {
	openedRepository, err := openStoredRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.deleteRepository()
}

func StorePending(repository *resources.Repository, ref string, args ...string) (err error) {
	openedRepository, err := openPrePushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.storePending(ref, repository.RemoteUri, args...)
}

func MergeChangeset(repository *resources.Repository, headRef, baseRef, refToMergeInto string) (err error) {
	openedRepository, err := openPrePushRepository(repository)
	if err != nil {
		return
	}

	lockMapMutex.Lock()
	repositoryLock, ok := lockMap[repository.Id]

	if !ok {
		repositoryLock = new(sync.Mutex)
		lockMap[repository.Id] = repositoryLock
	}

	lockMapMutex.Unlock()

	repositoryLock.Lock()
	defer repositoryLock.Unlock()

	return openedRepository.mergeChangeset(headRef, baseRef, refToMergeInto)
}
