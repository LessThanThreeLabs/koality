package database

import (
	"koality/resources"
	"testing"
)

const (
	repositoryName      = "repository-name"
	repositoryVcsType   = "git"
	repositoryLocalUri  = "git@local.uri.com:name.git"
	repositoryRemoteUri = "git@remote.uri.com:name.git"
	gitHubOwner         = "jordanpotter"
	gitHubName          = "repository-github-name"
	gitHubHookId        = 17
	gitHubHookSecret    = "hook-secret"
)

var (
	gitHubHookTypes []string = []string{"push", "pull_request"}
)

func TestCreateInvalidRepository(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	_, err = connection.Repositories.Create.Create("", repositoryVcsType, repositoryLocalUri, repositoryRemoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository name")
	}

	_, err = connection.Repositories.Create.Create("inv@lid-repos!tory-name^", repositoryVcsType, repositoryLocalUri, repositoryRemoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository name")
	}

	_, err = connection.Repositories.Create.Create(repositoryName, "", repositoryLocalUri, repositoryRemoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository VCS type")
	}

	_, err = connection.Repositories.Create.Create(repositoryName, "blah", repositoryLocalUri, repositoryRemoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository VCS type")
	}

	_, err = connection.Repositories.Create.Create(repositoryName, repositoryVcsType, "", repositoryRemoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository local uri")
	}

	_, err = connection.Repositories.Create.Create(repositoryName, repositoryVcsType, "google.com", repositoryRemoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository local uri")
	}

	_, err = connection.Repositories.Create.Create(repositoryName, repositoryVcsType, repositoryLocalUri, "")
	if err == nil {
		test.Fatal("Expected error after providing invalid repository remote uri")
	}

	_, err = connection.Repositories.Create.Create(repositoryName, repositoryVcsType, repositoryLocalUri, "google.com")
	if err == nil {
		test.Fatal("Expected error after providing invalid repository remote uri")
	}
}

func TestCreateRepository(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err := connection.Repositories.Create.Create(repositoryName, repositoryVcsType, repositoryLocalUri, repositoryRemoteUri)
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

	err = connection.Repositories.Update.SetGitHubHook(repositoryId, gitHubHookId, gitHubHookSecret, gitHubHookTypes)
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
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err := connection.Repositories.Create.CreateWithGitHub(repositoryName, repositoryVcsType, repositoryLocalUri, repositoryRemoteUri, gitHubOwner, gitHubName)
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

func TestGitHubHook(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err := connection.Repositories.Create.CreateWithGitHub(repositoryName, repositoryVcsType, repositoryLocalUri, repositoryRemoteUri, gitHubOwner, gitHubName)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Update.SetGitHubHook(repositoryId, gitHubHookId, gitHubHookSecret, gitHubHookTypes)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Update.ClearGitHubHook(repositoryId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Delete.Delete(repositoryId)
	if err != nil {
		test.Fatal(err)
	}
}
