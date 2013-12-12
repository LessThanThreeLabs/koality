package database

import (
	"koality/resources"
	"testing"
)

func TestCreateInvalidRepository(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	name := "repository-name"
	vcsType := "git"
	localUri := "git@local.uri.com:name.git"
	remoteUri := "git@remote.uri.com:name.git"

	_, err = connection.Repositories.Create.Create("", vcsType, localUri, remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository name")
	}

	_, err = connection.Repositories.Create.Create("inv@lid-repos!tory-name^", vcsType, localUri, remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository name")
	}

	_, err = connection.Repositories.Create.Create(name, "", localUri, remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository VCS type")
	}

	_, err = connection.Repositories.Create.Create(name, "blah", localUri, remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository VCS type")
	}

	_, err = connection.Repositories.Create.Create(name, vcsType, "", remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository local uri")
	}

	_, err = connection.Repositories.Create.Create(name, vcsType, "google.com", remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository local uri")
	}

	_, err = connection.Repositories.Create.Create(name, vcsType, localUri, "")
	if err == nil {
		test.Fatal("Expected error after providing invalid repository remote uri")
	}

	_, err = connection.Repositories.Create.Create(name, vcsType, localUri, "google.com")
	if err == nil {
		test.Fatal("Expected error after providing invalid repository remote uri")
	}
}

func TestCreateRepository(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err := connection.Repositories.Create.Create("repository-name", "hg", "hg@local.uri.com/name", "hg@remote.uri.com/name")
	if err != nil {
		test.Fatal(err)
	}

	repository, err := connection.Repositories.Read.Get(repositoryId)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Id != repositoryId {
		test.Fatal("repository.Id mismatch")
	}

	_, err = connection.Repositories.Create.Create(repository.Name, repository.VcsType, repository.LocalUri, repository.RemoteUri)
	if _, ok := err.(resources.RepositoryAlreadyExistsError); !ok {
		test.Fatal("Expected RepositoryAlreadyExistsError when trying to add same repository twice")
	}

	err = connection.Repositories.Update.SetGitHubHook(repositoryId, 17, "hook-secret", []string{"push", "pull_request"})
	if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
		test.Fatal("Expected NoSuchRepositoryHookError when trying to add repository hook")
	}

	err = connection.Repositories.Update.ClearGitHubHook(repositoryId)
	if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
		test.Fatal("Expected NoSuchRepositoryHookError when trying to clear repository hook")
	}

	err = connection.Repositories.Delete.Delete(repositoryId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Delete.Delete(repositoryId)
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when trying to delete same repository twice")
	}
}

func TestCreateGitHubRepository(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err := connection.Repositories.Create.CreateWithGitHub("repository-name", "git", "git@local.uri.com:name.git", "git@remote.uri.com:name.git", "jordanpotter", "repository-github-name")
	if err != nil {
		test.Fatal(err)
	}

	repository, err := connection.Repositories.Read.Get(repositoryId)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Id != repositoryId {
		test.Fatal("repository.Id mismatch")
	}

	_, err = connection.Repositories.Create.CreateWithGitHub(repository.Name, repository.VcsType, repository.LocalUri, repository.RemoteUri, repository.GitHub.Owner, repository.GitHub.Name)
	if _, ok := err.(resources.RepositoryAlreadyExistsError); !ok {
		test.Fatal("Expected RepositoryAlreadyExistsError when trying to add same repository twice")
	}

	err = connection.Repositories.Delete.Delete(repositoryId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Delete.Delete(repositoryId)
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when trying to delete same repository twice")
	}
}
