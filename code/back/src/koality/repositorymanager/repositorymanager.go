package repositorymanager

import (
	"koality/repositorymanager/repositorystore"
	"koality/resources"
)

type repoManager struct {
	repoType string
}

type RepoManager interface {
}

func GetYamlFile(repository *resources.Repository, ref string) (yamlFile string, err error) {
	vcsDispatcher := map[string]func(*resources.Repository, string) (string, error){
		"git":      getYamlFromGitRepo,
		"hg":       getYamlFromHgRepo,
		"perforce": getYamlFromPerforceRepo,
		"svn":      getYamlFromSvnRepo,
	}

	return vcsDispatcher[repository.VcsType](repository, ref)
}

func getYamlFromGitRepo(repository *resources.Repository, ref string) (yamlFile string, err error) {
	return repositorystore.GetYamlFile(repository, ref)
}

func getYamlFromHgRepo(repository *resources.Repository, ref string) (yamlFile string, err error) {
	return repositorystore.GetYamlFile(repository, ref)
}

func getYamlFromPerforceRepo(repository *resources.Repository, ref string) (yamlFile string, err error) {
	return
}
func getYamlFromSvnRepo(repository *resources.Repository, ref string) (yamlFile string, err error) {
	return
}

func GetCommitAttributes(repository *resources.Repository, ref string) (message, username, email string, err error) {
	vcsDispatcher := map[string]func(*resources.Repository, string) (string, string, string, error){
		"git":      getCommitAttributesFromGitRepo,
		"hg":       getCommitAttributesFromHgRepo,
		"perforce": getCommitAttributesFromPerforceRepo,
		"svn":      getCommitAttributesFromSvnRepo,
	}

	return vcsDispatcher[repository.VcsType](repository, ref)
}

func getCommitAttributesFromGitRepo(repository *resources.Repository, ref string) (message, username, email string, err error) {
	return repositorystore.GetCommitAttributes(repository, ref)
}

func getCommitAttributesFromHgRepo(repository *resources.Repository, ref string) (message, username, email string, err error) {
	return repositorystore.GetCommitAttributes(repository, ref)
}

func getCommitAttributesFromPerforceRepo(repository *resources.Repository, ref string) (message, username, email string, err error) {
	return
}

func getCommitAttributesFromSvnRepo(repository *resources.Repository, ref string) (message, username, email string, err error) {
	return
}
