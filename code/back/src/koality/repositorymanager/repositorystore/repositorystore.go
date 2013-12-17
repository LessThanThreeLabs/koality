package repositorystore

import (
	"koality/resources"
)

func CreateRepository(repo *resources.Repository) (err error) {
	return
}

func DeleteRepository(repo *resources.Repository) (err error) {
	return
}

func RenameRepository(repo *resources.Repository, oldName, newName string) (err error) {
	return
}

func MergeChangeset(repo *resources.Repository, headRef, baseRef, mergeIntoRef string) (err error) {
	return
}

func GetCommitAttributes(repo *resources.Repository, ref string) (message, username, email string, err error) {
	return
}

func GetYamlFile(repo *resources.Repository, ref string) (yamlFile string, err error) {
	return
}
