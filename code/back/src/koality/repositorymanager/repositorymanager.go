package repositorymanager

import (
	"fmt"
	"koality/resources"
	"sync"
)

var lockMap = make(map[uint64]sync.Locker)
var lockMapMutex sync.Mutex

type RepositoryManager interface {
	ToPath(repo *resources.Repository) string
	GetTopRef(repository *resources.Repository, branchOrRef string) (ref string, err error)
	GetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error)
	GetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error)
	CreateRepository(repository *resources.Repository) (err error)
	DeleteRepository(repository *resources.Repository) (err error)
	StorePending(repository *resources.Repository, ref string, args ...string) (err error)
	MergeChangeset(repository *resources.Repository, headRef, baseRef, refToMergeInto string) (err error)
}

type repositoryManager struct {
	path                string
	resourcesConnection *resources.Connection
}

func New(path string, resourcesConnection *resources.Connection) RepositoryManager {
	return &repositoryManager{path, resourcesConnection}
}

func (repositoryManager *repositoryManager) openPostPushRepository(repository *resources.Repository) (PostPushRepository, error) {
	switch repository.VcsType {
	case "git":
		return repositoryManager.openGitRepository(repository), nil
	case "hg":
		return repositoryManager.openHgRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}

}

func (repositoryManager *repositoryManager) openPrePushRepository(repository *resources.Repository) (PrePushRepository, error) {
	switch repository.VcsType {
	case "git":
		return repositoryManager.openGitRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}
}

func (repositoryManager *repositoryManager) openStoredRepository(repository *resources.Repository) (StoredRepository, error) {
	switch repository.VcsType {
	case "git":
		return repositoryManager.openGitRepository(repository), nil
	case "hg":
		return repositoryManager.openHgRepository(repository), nil
	default:
		return nil, fmt.Errorf("Repository type %s does not currently support the required post push operations.", repository.VcsType)
	}
}

func (repositoryManager *repositoryManager) GetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	openedRepository, err := repositoryManager.openPostPushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.getYamlFile(ref)
}

func (repositoryManager *repositoryManager) GetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	openedRepository, err := repositoryManager.openPostPushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.getCommitAttributes(ref)
}

func (repositoryManager *repositoryManager) CreateRepository(repository *resources.Repository) (err error) {
	openedRepository, err := repositoryManager.openStoredRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.createRepository()
}

func (repositoryManager *repositoryManager) DeleteRepository(repository *resources.Repository) (err error) {
	openedRepository, err := repositoryManager.openStoredRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.deleteRepository()
}

func (repositoryManager *repositoryManager) GetTopRef(repository *resources.Repository, branchOrRef string) (ref string, err error) {
	openedRepository, err := repositoryManager.openStoredRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.getTopRef(branchOrRef)
}

func (repositoryManager *repositoryManager) StorePending(repository *resources.Repository, ref string, args ...string) (err error) {
	openedRepository, err := repositoryManager.openPrePushRepository(repository)
	if err != nil {
		return
	}

	return openedRepository.storePending(ref, repository.RemoteUri, args...)
}

func (repositoryManager *repositoryManager) MergeChangeset(repository *resources.Repository, headRef, baseRef, refToMergeInto string) (err error) {
	openedRepository, err := repositoryManager.openPrePushRepository(repository)
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
