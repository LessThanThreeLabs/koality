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

func TestRepositoryStatus(test *testing.T) {
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
	} else if repository.Status != "" {
		test.Fatal("Expected initial repository status to be empty")
	}

	err = connection.Repositories.Update.SetStatus(repositoryId, "preparing")
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Update.SetStatus(repositoryId, "installed")
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Update.SetStatus(repositoryId, "bad-status")
	if _, ok := err.(resources.InvalidRepositoryStatusError); !ok {
		test.Fatal("Expected InvalidRepositoryStatusError when trying to set to invalid repository status")
	}

	err = connection.Repositories.Update.SetStatus(0, "installed")
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when trying to set to status for nonexistent repository")
	}
}

func TestRepositoryHook(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err := connection.Repositories.Create.CreateWithGitHub("repository-name", "git", "git@local.uri.com:name.git", "git@remote.uri.com:name.git", "jordanpotter", "repository-github-name")
	if err != nil {
		test.Fatal(err)
	}

	hookId := int64(17)
	hookSecret := "hook-secret"
	hookTypes := []string{"push", "pull_request"}

	err = connection.Repositories.Update.SetGitHubHook(repositoryId, hookId, hookSecret, hookTypes)
	if err != nil {
		test.Fatal(err)
	}

	repository, err := connection.Repositories.Read.Get(repositoryId)
	if err != nil {
		test.Fatal(err)
	} else if repository.GitHub.HookId != hookId {
		test.Fatal("Hook id not updated")
	} else if repository.GitHub.HookSecret != hookSecret {
		test.Fatal("Hook secret not updated")
	} else if len(repository.GitHub.HookTypes) != len(hookTypes) {
		test.Fatal("Hook types not updated")
	} else if repository.GitHub.HookTypes[0] != hookTypes[0] {
		test.Fatal("Hook types not updated")
	} else if repository.GitHub.HookTypes[1] != hookTypes[1] {
		test.Fatal("Hook types not updated")
	}

	err = connection.Repositories.Update.ClearGitHubHook(repositoryId)
	if err != nil {
		test.Fatal(err)
	}

	repository, err = connection.Repositories.Read.Get(repositoryId)
	if err != nil {
		test.Fatal(err)
	} else if repository.GitHub.HookId != 0 {
		test.Fatal("Hook id not updated")
	} else if repository.GitHub.HookSecret != "" {
		test.Fatal("Hook secret not updated")
	} else if len(repository.GitHub.HookTypes) != 0 {
		test.Fatal("Hook types not updated")
	}

	err = connection.Repositories.Update.SetGitHubHook(0, hookId, hookSecret, hookTypes)
	if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
		test.Fatal("Expected NoSuchRepositoryHookError when trying to set hook for nonexistent repository")
	}

	err = connection.Repositories.Update.ClearGitHubHook(0)
	if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
		test.Fatal("Expected NoSuchRepositoryHookError when trying to clear hook for nonexistent repository")
	}
}
