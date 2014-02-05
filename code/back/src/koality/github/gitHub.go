package github

import (
	"fmt"
	"github.com/LessThanThreeLabs/goauth2/oauth"
	"github.com/google/go-github/github"
	"koality/resources"
	"net/url"
)

type GitHubRepository struct {
	Owner string
	Name  string
}

type GitHubConnection interface {
	GetSshKeys(oAuthToken string) (sshKeys []string, err error)
	AddKoalitySshKeyToUser(oAuthToken string) error
	GetRepositories(oAuthToken string) (repositories []GitHubRepository, err error)
	AddRepositoryHook(repository *resources.Repository, hookTypes []string, hookSecret string) (hookId int64, err error)
	RemoveRepositoryHook(repository *resources.Repository) error
	SetChangeStatus(repository *resources.Repository, sha, status, description, url string) error
}

type gitHubConnection struct {
	resourcesConnection *resources.Connection
	oAuthConnection     GitHubOAuthConnection
}

func NewConnection(resourcesConnection *resources.Connection, oAuthConnection GitHubOAuthConnection) *gitHubConnection {
	return &gitHubConnection{resourcesConnection, oAuthConnection}
}

// TODO (bbland): there's no way this works
func (connection *gitHubConnection) getGitHubClient(oAuthToken string) (*github.Client, error) {
	gitHubEnterpriseSettings, err := connection.resourcesConnection.Settings.Read.GetGitHubEnterpriseSettings()
	if _, ok := err.(resources.NoSuchSettingError); err != nil && !ok {
		return nil, err
	}

	ok, err := connection.oAuthConnection.CheckValidOAuthToken(oAuthToken)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("Invalid OAuth token: %s", oAuthToken)
	}

	transport := &oauth.Transport{
		Token: &oauth.Token{AccessToken: oAuthToken},
	}
	gitHubClient := github.NewClient(transport.Client())
	if gitHubEnterpriseSettings != nil {
		baseUrl, err := url.Parse(gitHubEnterpriseSettings.BaseUrl)
		if err != nil {
			return nil, err
		}
		gitHubClient.BaseURL = baseUrl
	}
	return gitHubClient, nil
}

func (connection *gitHubConnection) GetSshKeys(oAuthToken string) ([]string, error) {
	gitHubClient, err := connection.getGitHubClient(oAuthToken)
	if err != nil {
		return nil, err
	}

	keys, _, err := gitHubClient.Users.ListKeys("")
	if err != nil {
		return nil, err
	}

	sshKeys := make([]string, len(keys))
	for index, key := range keys {
		sshKeys[index] = *key.Key
	}

	return sshKeys, nil
}

func (connection *gitHubConnection) AddKoalitySshKeyToUser(oAuthToken string) error {
	gitHubClient, err := connection.getGitHubClient(oAuthToken)
	if err != nil {
		return err
	}

	keyPair, err := connection.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		return err
	}

	_, _, err = gitHubClient.Users.CreateKey(&github.Key{
		Title: github.String("Koality Build"),
		Key:   &keyPair.PublicKey,
	})
	return err
}

func (connection *gitHubConnection) GetRepositories(oAuthToken string) ([]GitHubRepository, error) {
	gitHubClient, err := connection.getGitHubClient(oAuthToken)
	if err != nil {
		return nil, err
	}

	gitHubUser, _, err := gitHubClient.Users.Get("")
	if err != nil {
		return nil, err
	}

	organizations, _, err := gitHubClient.Organizations.List("", nil)
	if err != nil {
		return nil, err
	}

	repositories, _, err := gitHubClient.Repositories.List(*gitHubUser.Name, nil)
	if err != nil {
		return nil, err
	}

	for _, organization := range organizations {
		orgRepositories, _, err := gitHubClient.Repositories.ListByOrg(*organization.Name, nil)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, orgRepositories...)
	}

	gitHubRepositories := make([]GitHubRepository, len(repositories))
	for index, repository := range repositories {
		gitHubRepositories[index] = GitHubRepository{*repository.Owner.Name, *repository.Name}
	}
	return gitHubRepositories, nil
}

func (connection *gitHubConnection) AddRepositoryHook(repository *resources.Repository, hookTypes []string, hookSecret string) (int64, error) {
	gitHubClient, err := connection.getGitHubClient(repository.GitHub.OAuthToken)
	if err != nil {
		return 0, err
	}

	hook := github.Hook{
		Name:   github.String("web"),
		Events: hookTypes,
		Active: github.Bool(true),
		Config: map[string]interface{}{
			"secret":       hookSecret,
			"insecure_ssl": 1,
		},
	}

	createdHook, _, err := gitHubClient.Repositories.CreateHook(repository.GitHub.Owner, repository.GitHub.Name, &hook)
	if err != nil {
		return 0, err
	}

	return int64(*createdHook.ID), nil
}

func (connection *gitHubConnection) RemoveRepositoryHook(repository *resources.Repository) error {
	gitHubClient, err := connection.getGitHubClient(repository.GitHub.OAuthToken)
	if err != nil {
		return err
	}

	_, err = gitHubClient.Repositories.DeleteHook(repository.GitHub.Owner, repository.GitHub.Name, int(repository.GitHub.HookId))
	return err
}

func (connection *gitHubConnection) SetChangeStatus(repository *resources.Repository, sha, status, description, url string) error {
	gitHubClient, err := connection.getGitHubClient(repository.GitHub.OAuthToken)
	if err != nil {
		return err
	}

	repoStatus := github.RepoStatus{
		State:       &status,
		Description: &description,
		TargetURL:   &url,
	}

	_, _, err = gitHubClient.Repositories.CreateStatus(repository.GitHub.Owner, repository.GitHub.Name, sha, &repoStatus)
	return err
}
