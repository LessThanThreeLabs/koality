package repomanager

import (
	"koality/repomanager/repostore"
	"koality/resources"
)

type repoManager struct {
	repoType string
}

type RepoManager interface {
}

func GetYamlFile(repo *resources.Repository, ref string) (yamlFile string, err error) {
	vcsDispatcher := map[string]func(*resources.Repository, string) (string, error){
		"git":      pullYamlFromGitRepo,
		"hg":       pullYamlFromHgRepo,
		"perforce": pullYamlFromPerforceRepo,
		"svn":      pullYamlFromSvnRepo,
	}

	return vcsDispatcher[repo.VcsType](repo, ref)
}

func pullYamlFromGitRepo(repo *resources.Repository, ref string) (yamlFile string, err error) {
	return
}

func pullYamlFromHgRepo(repo *resources.Repository, ref string) (yamlFile string, err error) {
	return
}

func pullYamlFromPerforceRepo(repo *resources.Repository, ref string) (yamlFile string, err error) {
	return
}
func pullYamlFromSvnRepo(repo *resources.Repository, ref string) (yamlFile string, err error) {
	return
}

func GetCommitAttributes(repo *resources.Repository, ref string) (message, username, email string, err error) {
	vcsDispatcher := map[string]func(*resources.Repository, string) (string, string, string, error){
		"git":      getCommitAttributesFromGitRepo,
		"hg":       getCommitAttributesFromHgRepo,
		"perforce": getCommitAttributesFromPerforceRepo,
		"svn":      getCommitAttributesFromSvnRepo,
	}

	return vcsDispatcher[repo.VcsType](repo, ref)
}

func getCommitAttributesFromGitRepo(repo *resources.Repository, ref string) (message, username, email string, err error) {
	return
}

func getCommitAttributesFromHgRepo(repo *resources.Repository, ref string) (message, username, email string, err error) {
	return
}

func getCommitAttributesFromPerforceRepo(repo *resources.Repository, ref string) (message, username, email string, err error) {
	return
}

func getCommitAttributesFromSvnRepo(repo *resources.Repository, ref string) (message, username, email string, err error) {
	return
}
