package github

import (
	"github.com/LessThanThreeLabs/goauth2/oauth"
	"github.com/google/go-github/github"
	"koality/resources"
	"net/url"
)

type GitHubRepository struct {
	Owner string
	Name  string
}

type GitHubClient struct {
	resourcesConnection *resources.Connection
}

func NewClient(resourcesConnection *resources.Connection) *GitHubClient {
	return &GitHubClient{resourcesConnection}
}

// TODO (bbland): there's no way this works
func (client *GitHubClient) getGitHubClient(oAuthToken string) (*github.Client, error) {
	gitHubEnterpriseSettings, err := client.resourcesConnection.Settings.Read.GetGitHubEnterpriseSettings()

	var oAuthConfig *oauth.Config
	if _, ok := err.(resources.NoSuchSettingError); err != nil && !ok {
		return nil, err
	} else if err == nil {
		oAuthConfig = &oauth.Config{
			ClientId:     gitHubEnterpriseSettings.OAuthClientId,
			ClientSecret: gitHubEnterpriseSettings.OAuthClientSecret,
		}
	}

	transport := &oauth.Transport{
		Config: oAuthConfig,
		Token:  &oauth.Token{AccessToken: oAuthToken},
	}
	gitHubClient := github.NewClient(transport.Client())
	if gitHubEnterpriseSettings != nil {
		baseUrl, err := url.Parse(gitHubEnterpriseSettings.Url)
		if err != nil {
			return nil, err
		}
		gitHubClient.BaseURL = baseUrl
	}
	return gitHubClient, nil
}

func (client *GitHubClient) GetSshKeys(user resources.User) ([]string, error) {
	gitHubClient, err := client.getGitHubClient(user.GitHubOAuth)
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

func (client *GitHubClient) AddKoalitySshKeyToUser(user resources.User) error {
	gitHubClient, err := client.getGitHubClient(user.GitHubOAuth)
	if err != nil {
		return err
	}

	keyPair, err := client.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		return err
	}

	_, _, err = gitHubClient.Users.CreateKey(&github.Key{
		Title: github.String("Koality Verification"),
		Key:   &keyPair.PublicKey,
	})
	return err
}

func (client *GitHubClient) GetRepositories(user resources.User) ([]GitHubRepository, error) {
	gitHubClient, err := client.getGitHubClient(user.GitHubOAuth)
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

func (client *GitHubClient) AddRepositoryHook(user resources.User, repository resources.Repository, hookTypes []string, hookSecret string) (int64, error) {
	gitHubClient, err := client.getGitHubClient(user.GitHubOAuth)
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

func (client *GitHubClient) RemoveRepositoryHook(user resources.User, repository resources.Repository) error {
	gitHubClient, err := client.getGitHubClient(user.GitHubOAuth)
	if err != nil {
		return err
	}

	_, err = gitHubClient.Repositories.DeleteHook(repository.GitHub.Owner, repository.GitHub.Name, int(repository.GitHub.HookId))
	return err
}
