package database

import (
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidRepository(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	name := "repository-name"
	vcsType := "git"
	localUri := "git/local_uri/name.git"
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

	_, err = connection.Repositories.Create.Create(name, vcsType, "git@google.com", remoteUri)
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

	createdEventReceived := make(chan bool, 1)
	var createdEventRepository *resources.Repository
	repositoryCreatedHandler := func(repository *resources.Repository) {
		createdEventRepository = repository
		createdEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToCreatedEvents(repositoryCreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	deletedEventReceived := make(chan bool, 1)
	deletedEventId := uint64(0)
	repositoryDeletedHandler := func(repositoryId uint64) {
		deletedEventId = repositoryId
		deletedEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToDeletedEvents(repositoryDeletedHandler)
	if err != nil {
		test.Fatal(err)
	}

	name := "repository-name"
	vcsType := "hg"
	localUri := "hg/local_uri/name"
	remoteUri := "hg@remote.uri.com/name"
	repository, err := connection.Repositories.Create.Create(name, vcsType, localUri, remoteUri)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Name != name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != vcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.LocalUri != localUri {
		test.Fatal("repository.LocalUri mismatch")
	} else if repository.RemoteUri != remoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub != nil {
		test.Fatal("Expected repository.GitHub to be nil")
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-createdEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository creation event")
	}

	if createdEventRepository.Id != repository.Id {
		test.Fatal("Bad repository.Id in repository creation event")
	} else if createdEventRepository.Name != repository.Name {
		test.Fatal("Bad repository.Name in repository creation event")
	} else if createdEventRepository.VcsType != repository.VcsType {
		test.Fatal("Bad repository.VcsType in repository creation event")
	} else if createdEventRepository.LocalUri != repository.LocalUri {
		test.Fatal("Bad repository.LocalUri in repository creation event")
	} else if createdEventRepository.RemoteUri != repository.RemoteUri {
		test.Fatal("Bad repository.RemoteUri in repository creation event")
	} else if createdEventRepository.GitHub != repository.GitHub {
		test.Fatal("Bad repository.GitHub in repository creation event")
	}

	repository2, err := connection.Repositories.Read.Get(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Id != repository2.Id {
		test.Fatal("repository.Id mismatch")
	} else if repository.Name != repository2.Name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != repository2.VcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.LocalUri != repository2.LocalUri {
		test.Fatal("repository.LocalUri mismatch")
	} else if repository.RemoteUri != repository2.RemoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub != repository2.GitHub {
		test.Fatal("repository.GitHub mismatch")
	}

	_, err = connection.Repositories.Create.Create(repository.Name, repository.VcsType, repository.LocalUri, repository.RemoteUri)
	if _, ok := err.(resources.RepositoryAlreadyExistsError); !ok {
		test.Fatal("Expected RepositoryAlreadyExistsError when trying to add same repository twice")
	}

	err = connection.Repositories.Update.SetGitHubHook(repository.Id, 17, "hook-secret", []string{"push", "pull_request"})
	if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
		test.Fatal("Expected NoSuchRepositoryHookError when trying to add repository hook")
	}

	err = connection.Repositories.Update.ClearGitHubHook(repository.Id)
	if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
		test.Fatal("Expected NoSuchRepositoryHookError when trying to clear repository hook")
	}

	err = connection.Repositories.Delete.Delete(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-deletedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository deletion event")
	}

	if deletedEventId != repository.Id {
		test.Fatal("Bad repository.Id in repository deletion event")
	}

	err = connection.Repositories.Delete.Delete(repository.Id)
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

	createdEventReceived := make(chan bool, 1)
	var createdEventRepository *resources.Repository
	repositoryCreatedHandler := func(repository *resources.Repository) {
		createdEventRepository = repository
		createdEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToCreatedEvents(repositoryCreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	deletedEventReceived := make(chan bool, 1)
	deletedEventId := uint64(0)
	repositoryDeletedHandler := func(repositoryId uint64) {
		deletedEventId = repositoryId
		deletedEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToDeletedEvents(repositoryDeletedHandler)
	if err != nil {
		test.Fatal(err)
	}

	name := "repository-name"
	vcsType := "git"
	localUri := "git/local.uri.com/name"
	remoteUri := "git@remote_uri:name"
	gitHubOwner := "jordanpotter"
	gitHubName := "repository-github-name"
	repository, err := connection.Repositories.Create.CreateWithGitHub(name, vcsType, localUri, remoteUri, gitHubOwner, gitHubName)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Name != name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != vcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.LocalUri != localUri {
		test.Fatal("repository.LocalUri mismatch")
	} else if repository.RemoteUri != remoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != gitHubOwner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != gitHubName {
		test.Fatal("repository.GitHub.Name mismatch")
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-createdEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository creation event")
	}

	if createdEventRepository.Id != repository.Id {
		test.Fatal("Bad repository.Id in repository creation event")
	} else if createdEventRepository.Name != repository.Name {
		test.Fatal("Bad repository.Name in repository creation event")
	} else if createdEventRepository.VcsType != repository.VcsType {
		test.Fatal("Bad repository.VcsType in repository creation event")
	} else if createdEventRepository.LocalUri != repository.LocalUri {
		test.Fatal("Bad repository.LocalUri in repository creation event")
	} else if createdEventRepository.RemoteUri != repository.RemoteUri {
		test.Fatal("Bad repository.RemoteUri in repository creation event")
	} else if createdEventRepository.GitHub.Owner != repository.GitHub.Owner {
		test.Fatal("Bad repository.GitHub.Owner in repository creation event")
	} else if createdEventRepository.GitHub.Name != repository.GitHub.Name {
		test.Fatal("Bad repository.GitHub.Name in repository creation event")
	}

	repository2, err := connection.Repositories.Read.Get(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Id != repository2.Id {
		test.Fatal("repository.Id mismatch")
	} else if repository.Name != repository2.Name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != repository2.VcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.LocalUri != repository2.LocalUri {
		test.Fatal("repository.LocalUri mismatch")
	} else if repository.RemoteUri != repository2.RemoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != repository2.GitHub.Owner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != repository2.GitHub.Name {
		test.Fatal("repository.GitHub.Name mismatch")
	}

	_, err = connection.Repositories.Create.CreateWithGitHub(repository.Name, repository.VcsType, repository.LocalUri, repository.RemoteUri, repository.GitHub.Owner, repository.GitHub.Name)
	if _, ok := err.(resources.RepositoryAlreadyExistsError); !ok {
		test.Fatal("Expected RepositoryAlreadyExistsError when trying to add same repository twice")
	}

	err = connection.Repositories.Delete.Delete(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-deletedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository deletion event")
	}

	if deletedEventId != repository.Id {
		test.Fatal("Bad repositoryId in repository deletion event")
	}

	err = connection.Repositories.Delete.Delete(repository.Id)
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

	repositoryEventReceived := make(chan bool, 1)
	repositoryEventId := uint64(0)
	repositoryEventStatus := ""
	repositoryStatusUpdatedHandler := func(repositoryId uint64, status string) {
		repositoryEventId = repositoryId
		repositoryEventStatus = status
		repositoryEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToStatusUpdatedEvents(repositoryStatusUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	repository, err := connection.Repositories.Create.Create("repository-name", "hg", "hg/local_uri/name", "hg@remote.uri.com/name")
	if err != nil {
		test.Fatal(err)
	}

	if repository.Status != "declared" {
		test.Fatal("Expected initial repository status to be 'declared'")
	}

	err = connection.Repositories.Update.SetStatus(repository.Id, "preparing")
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-repositoryEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository status updated event")
	}

	if repositoryEventId != repository.Id {
		test.Fatal("Bad repositoryId in status updated event")
	} else if repositoryEventStatus != "preparing" {
		test.Fatal("Bad repository status in status updated event")
	}

	err = connection.Repositories.Update.SetStatus(repository.Id, "installed")
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-repositoryEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository status event")
	}

	if repositoryEventId != repository.Id {
		test.Fatal("Bad repositoryId in status event")
	} else if repositoryEventStatus != "installed" {
		test.Fatal("Bad repository status in status event")
	}

	err = connection.Repositories.Update.SetStatus(repository.Id, "bad-status")
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

	repositoryEventReceived := make(chan bool, 1)
	repositoryEventId := uint64(0)
	repositoryEventHookId := int64(0)
	repositoryEventHookSecret := ""
	repositoryEventHookTypes := []string{}
	repositoryGitHubHookUpdatedHandler := func(repositoryId uint64, hookId int64, hookSecret string, hookTypes []string) {
		repositoryEventId = repositoryId
		repositoryEventHookId = hookId
		repositoryEventHookSecret = hookSecret
		repositoryEventHookTypes = hookTypes
		repositoryEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToGitHubHookUpdatedEvents(repositoryGitHubHookUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	repository, err := connection.Repositories.Create.CreateWithGitHub("repository-name", "git", "git/local_uri/name.git", "git@remote.uri.com:name.git", "jordanpotter", "repository-github-name")
	if err != nil {
		test.Fatal(err)
	}

	hookId := int64(17)
	hookSecret := "hook-secret"
	hookTypes := []string{"push", "pull_request"}

	err = connection.Repositories.Update.SetGitHubHook(repository.Id, hookId, hookSecret, hookTypes)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-repositoryEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository hook event")
	}

	if repositoryEventId != repository.Id {
		test.Fatal("Bad repositoryId in hook updated event")
	} else if repositoryEventHookId != hookId {
		test.Fatal("Bad repository hook id in hook updated event")
	} else if repositoryEventHookSecret != hookSecret {
		test.Fatal("Bad repository hook secret in hook updated event")
	} else if len(repositoryEventHookTypes) != len(hookTypes) {
		test.Fatal("Bad repository hook types in hook updated event")
	} else if repositoryEventHookTypes[0] != hookTypes[0] {
		test.Fatal("Bad repository hook types in hook updated event")
	} else if repositoryEventHookTypes[1] != hookTypes[1] {
		test.Fatal("Bad repository hook types in hook updated event")
	}

	repository2, err := connection.Repositories.Read.Get(repository.Id)
	if err != nil {
		test.Fatal(err)
	} else if repository2.GitHub.HookId != hookId {
		test.Fatal("Hook id not updated")
	} else if repository2.GitHub.HookSecret != hookSecret {
		test.Fatal("Hook secret not updated")
	} else if len(repository2.GitHub.HookTypes) != len(hookTypes) {
		test.Fatal("Hook types not updated")
	} else if repository2.GitHub.HookTypes[0] != hookTypes[0] {
		test.Fatal("Hook types not updated")
	} else if repository2.GitHub.HookTypes[1] != hookTypes[1] {
		test.Fatal("Hook types not updated")
	}

	err = connection.Repositories.Update.ClearGitHubHook(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-repositoryEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear repository hook event")
	}

	if repositoryEventId != repository.Id {
		test.Fatal("Bad repositoryId in hook updated event")
	} else if repositoryEventHookId != 0 {
		test.Fatal("Bad repository hook id in hook updated event")
	} else if repositoryEventHookSecret != "" {
		test.Fatal("Bad repository hook secret in hook updated event")
	} else if len(repositoryEventHookTypes) != 0 {
		test.Fatal("Bad repository hook types in hook updated event")
	}

	repository2, err = connection.Repositories.Read.Get(repository2.Id)
	if err != nil {
		test.Fatal(err)
	} else if repository2.GitHub.HookId != 0 {
		test.Fatal("Hook id not updated")
	} else if repository2.GitHub.HookSecret != "" {
		test.Fatal("Hook secret not updated")
	} else if len(repository2.GitHub.HookTypes) != 0 {
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
