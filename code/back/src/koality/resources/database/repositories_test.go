package database

import (
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidRepository(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	name := "repository-name"
	vcsType := "git"
	remoteUri := "git@remote.uri.com:name.git"

	_, err = connection.Repositories.Create.Create("", vcsType, remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository name")
	}

	_, err = connection.Repositories.Create.Create("inv@lid-repos!tory-name^", vcsType, remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository name")
	}

	_, err = connection.Repositories.Create.Create(name, "", remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository VCS type")
	}

	_, err = connection.Repositories.Create.Create(name, "blah", remoteUri)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository VCS type")
	}

	_, err = connection.Repositories.Create.Create(name, vcsType, "")
	if err == nil {
		test.Fatal("Expected error after providing invalid repository remote uri")
	}

	_, err = connection.Repositories.Create.Create(name, vcsType, "google.com")
	if err == nil {
		test.Fatal("Expected error after providing invalid repository remote uri")
	}
}

func TestCreateAndDeleteRepository(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

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
	remoteUri := "hg@remote.uri.com/name"
	repository, err := connection.Repositories.Create.Create(name, vcsType, remoteUri)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Name != name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != vcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.RemoteUri != remoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub != nil {
		test.Fatal("Expected repository.GitHub to be nil")
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear repository creation event")
	}

	if createdEventRepository.Id != repository.Id {
		test.Fatal("Bad repository.Id in repository creation event")
	} else if createdEventRepository.Name != repository.Name {
		test.Fatal("Bad repository.Name in repository creation event")
	} else if createdEventRepository.VcsType != repository.VcsType {
		test.Fatal("Bad repository.VcsType in repository creation event")
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
	} else if repository.RemoteUri != repository2.RemoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub != repository2.GitHub {
		test.Fatal("repository.GitHub mismatch")
	}

	_, err = connection.Repositories.Create.Create(repository.Name, repository.VcsType, repository.RemoteUri)
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

	select {
	case <-deletedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear repository deletion event")
	}

	if deletedEventId != repository.Id {
		test.Fatal("Bad repository.Id in repository deletion event")
	}

	err = connection.Repositories.Delete.Delete(repository.Id)
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when trying to delete same repository twice")
	}

	deletedRepository, err := connection.Repositories.Read.Get(repository.Id)
	if err != nil {
		test.Fatal(err)
	} else if !deletedRepository.IsDeleted {
		test.Fatal("Expected repository to be marked as deleted")
	}
}

func TestCreateGitHubRepository(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

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
	remoteUri := "git@remote_uri:name"
	gitHubOwner := "jordanpotter"
	gitHubName := "repository-github-name"
	oAuthToken := "ABC123"
	repository, err := connection.Repositories.Create.CreateWithGitHub(name, remoteUri, gitHubOwner, gitHubName, oAuthToken)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Name != name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != vcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.RemoteUri != remoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != gitHubOwner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != gitHubName {
		test.Fatal("repository.GitHub.Name mismatch")
	} else if repository.GitHub.OAuthToken != oAuthToken {
		test.Fatal("repository.GitHub.OAuthToken mismatch")
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear repository creation event")
	}

	if createdEventRepository.Id != repository.Id {
		test.Fatal("Bad repository.Id in repository creation event")
	} else if createdEventRepository.Name != repository.Name {
		test.Fatal("Bad repository.Name in repository creation event")
	} else if createdEventRepository.VcsType != repository.VcsType {
		test.Fatal("Bad repository.VcsType in repository creation event")
	} else if createdEventRepository.RemoteUri != repository.RemoteUri {
		test.Fatal("Bad repository.RemoteUri in repository creation event")
	} else if createdEventRepository.GitHub.Owner != repository.GitHub.Owner {
		test.Fatal("Bad repository.GitHub.Owner in repository creation event")
	} else if createdEventRepository.GitHub.Name != repository.GitHub.Name {
		test.Fatal("Bad repository.GitHub.Name in repository creation event")
	} else if createdEventRepository.GitHub.OAuthToken != repository.GitHub.OAuthToken {
		test.Fatal("Bad repository.GitHub.OAuthToken in repository creation event")
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
	} else if repository.RemoteUri != repository2.RemoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != repository2.GitHub.Owner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != repository2.GitHub.Name {
		test.Fatal("repository.GitHub.Name mismatch")
	} else if repository.GitHub.OAuthToken != repository2.GitHub.OAuthToken {
		test.Fatal("repository.GitHub.OAuthToken mismatch")
	}

	repository3, err := connection.Repositories.Read.GetByGitHubInfo(repository.GitHub.Owner, repository.GitHub.Name)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Id != repository3.Id {
		test.Fatal("repository.Id mismatch")
	} else if repository.Name != repository3.Name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != repository3.VcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.RemoteUri != repository3.RemoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != repository3.GitHub.Owner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != repository3.GitHub.Name {
		test.Fatal("repository.GitHub.Name mismatch")
	} else if repository.GitHub.OAuthToken != repository3.GitHub.OAuthToken {
		test.Fatal("repository.GitHub.OAuthToken mismatch")
	}

	_, err = connection.Repositories.Create.CreateWithGitHub(repository.Name, repository.RemoteUri, repository.GitHub.Owner, repository.GitHub.Name, "A different OAuth token")
	if _, ok := err.(resources.RepositoryAlreadyExistsError); !ok {
		test.Fatal("Expected RepositoryAlreadyExistsError when trying to add same repository twice")
	}

	err = connection.Repositories.Delete.Delete(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-deletedEventReceived:
	case <-time.After(10 * time.Second):
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

func TestRepositoryRemoteUri(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	repositoryEventReceived := make(chan bool, 1)
	repositoryEventId := uint64(0)
	repositoryEventRemoteUri := ""
	repositoryRemoteUriUpdatedHandler := func(repositoryId uint64, remoteUri string) {
		repositoryEventId = repositoryId
		repositoryEventRemoteUri = remoteUri
		repositoryEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToRemoteUriUpdatedEvents(repositoryRemoteUriUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	remoteUri1 := "hg@remote.uri.com/name"
	repository, err := connection.Repositories.Create.Create("repository-name", "hg", remoteUri1)
	if err != nil {
		test.Fatal(err)
	}

	remoteUri2 := "hg@remote.uri.com/name2"
	err = connection.Repositories.Update.SetRemoteUri(repository.Id, remoteUri2)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-repositoryEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear repository remote uri updated event")
	}

	if repositoryEventId != repository.Id {
		test.Fatal("Bad repositoryId in remote uri updated event")
	} else if repositoryEventRemoteUri != remoteUri2 {
		test.Fatal("Bad repository remote uri in remote uri updated event")
	}

	remoteUri3 := "hg@remote.uri.com/name3"
	err = connection.Repositories.Update.SetRemoteUri(repository.Id, remoteUri3)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-repositoryEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear remote uri event")
	}

	if repositoryEventId != repository.Id {
		test.Fatal("Bad repositoryId in remote uri event")
	} else if repositoryEventRemoteUri != remoteUri3 {
		test.Fatal("Bad repository remote uri in remote uri event")
	}

	err = connection.Repositories.Update.SetRemoteUri(repository.Id, "bad-remote-uri")
	if _, ok := err.(resources.InvalidRepositoryRemoteUriError); !ok {
		test.Fatal("Expected InvalidRepositoryRemoteUriError when trying to set to invalid repository remote uri")
	}

	err = connection.Repositories.Update.SetRemoteUri(0, "installed")
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when trying to set to remote uri for nonexistent repository")
	}
}

func TestRepositoryStatus(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

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

	repository, err := connection.Repositories.Create.Create("repository-name", "hg", "hg@remote.uri.com/name")
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

	select {
	case <-repositoryEventReceived:
	case <-time.After(10 * time.Second):
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

	select {
	case <-repositoryEventReceived:
	case <-time.After(10 * time.Second):
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

func TestGitHubOAuthTokens(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	oAuthUpdatedEventReceived := make(chan bool, 1)
	var oAuthEventRepositoryId uint64
	var oAuthEventToken string
	repositoryGitHubOAuthTokenUpdatedHandler := func(repositoryId uint64, oAuthToken string) {
		oAuthEventRepositoryId = repositoryId
		oAuthEventToken = oAuthToken
		oAuthUpdatedEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToGitHubOAuthTokenUpdatedEvents(repositoryGitHubOAuthTokenUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	oAuthClearedEventReceived := make(chan bool, 1)
	var oAuthClearedRepositoryId uint64
	repositoryGitHubOAuthTokenClearedHandler := func(repositoryId uint64) {
		oAuthClearedRepositoryId = repositoryId
		oAuthClearedEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToGitHubOAuthTokenClearedEvents(repositoryGitHubOAuthTokenClearedHandler)
	if err != nil {
		test.Fatal(err)
	}

	name := "repository-name"
	vcsType := "git"
	remoteUri := "git@remote_uri:name"
	gitHubOwner := "jordanpotter"
	gitHubName := "repository-github-name"
	oAuthToken := "ABC123"
	repository, err := connection.Repositories.Create.CreateWithGitHub(name, remoteUri, gitHubOwner, gitHubName, oAuthToken)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Name != name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != vcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.RemoteUri != remoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != gitHubOwner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != gitHubName {
		test.Fatal("repository.GitHub.Name mismatch")
	} else if repository.GitHub.OAuthToken != oAuthToken {
		test.Fatal("repository.GitHub.OAuthToken mismatch")
	}

	newOAuthToken := "A new OAuth Token"
	err = connection.Repositories.Update.SetGitHubOAuthToken(repository.Id, newOAuthToken)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-oAuthUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear oAuthToken updated event")
	}

	if oAuthEventRepositoryId != repository.Id {
		test.Fatal("Bad repositoryId in oAuthToken updated event")
	} else if oAuthEventToken != newOAuthToken {
		test.Fatal("Bad oAuthToken in oAuthToken updated event")
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
	} else if repository.RemoteUri != repository2.RemoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != repository2.GitHub.Owner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != repository2.GitHub.Name {
		test.Fatal("repository.GitHub.Name mismatch")
	} else if repository2.GitHub.OAuthToken != newOAuthToken {
		test.Fatal("repository.GitHub.OAuthToken was not updated")
	}

	err = connection.Repositories.Update.ClearGitHubOAuthToken(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-oAuthClearedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear oAuthToken cleared event")
	}

	if oAuthClearedRepositoryId != repository.Id {
		test.Fatal("Bad repositoryId in oAuthToken cleared event")
	}

	repository3, err := connection.Repositories.Read.Get(repository.Id)
	if err != nil {
		test.Fatal(err)
	}

	if repository.Id != repository3.Id {
		test.Fatal("repository.Id mismatch")
	} else if repository.Name != repository3.Name {
		test.Fatal("repository.Name mismatch")
	} else if repository.VcsType != repository3.VcsType {
		test.Fatal("repository.VcsType mismatch")
	} else if repository.RemoteUri != repository3.RemoteUri {
		test.Fatal("repository.RemoteUri mismatch")
	} else if repository.GitHub.Owner != repository3.GitHub.Owner {
		test.Fatal("repository.GitHub.Owner mismatch")
	} else if repository.GitHub.Name != repository3.GitHub.Name {
		test.Fatal("repository.GitHub.Name mismatch")
	} else if repository3.GitHub.OAuthToken != "" {
		test.Fatal("repository.GitHub.OAuthToken was not cleared")
	}
}

func TestRepositoryHook(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	repositoryEventReceived := make(chan bool, 1)
	var repositoryEventId uint64
	var repositoryEventHookId int64
	var repositoryEventHookSecret string
	var repositoryEventHookTypes []string
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

	repositoryHookClearedEventReceived := make(chan bool, 1)
	var repositoryHookClearedId uint64
	repositoryGitHubHookClearedHandler := func(repositoryId uint64) {
		repositoryHookClearedId = repositoryId
		repositoryHookClearedEventReceived <- true
	}
	_, err = connection.Repositories.Subscription.SubscribeToGitHubHookClearedEvents(repositoryGitHubHookClearedHandler)
	if err != nil {
		test.Fatal(err)
	}

	repository, err := connection.Repositories.Create.CreateWithGitHub("repository-name", "git@remote.uri.com:name.git", "jordanpotter", "repository-github-name", "oAuthToken")
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

	select {
	case <-repositoryEventReceived:
	case <-time.After(10 * time.Second):
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

	select {
	case <-repositoryHookClearedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear repository hook cleared event")
	}

	if repositoryHookClearedId != repository.Id {
		test.Fatal("Bad repositoryId in hook cleared event")
	}

	repository3, err := connection.Repositories.Read.Get(repository.Id)
	if err != nil {
		test.Fatal(err)
	} else if repository3.GitHub.HookId != 0 {
		test.Fatal("Hook id not cleared")
	} else if repository3.GitHub.HookSecret != "" {
		test.Fatal("Hook secret not cleared")
	} else if len(repository3.GitHub.HookTypes) != 0 {
		test.Fatal("Hook types not cleared")
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
