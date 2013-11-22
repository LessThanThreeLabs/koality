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

	err = connection.Repositories.Delete.Delete(repositoryId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Repositories.Delete.Delete(repositoryId)
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when trying to delete same repository twice")
	}
}
