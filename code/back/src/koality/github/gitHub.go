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
func (client *GitHubClient) getGitHubClient(oauthToken string) (*github.Client, error) {
	gitHubEnterpriseSettings, err := client.resourcesConnection.Settings.Read.GetGitHubEnterpriseSettings()

	var oauthConfig *oauth.Config
	if _, ok := err.(resources.NoSuchSettingError); err != nil && !ok {
		return nil, err
	} else if err == nil {
		oauthConfig = &oauth.Config{
			ClientId:     gitHubEnterpriseSettings.OAuthClientId,
			ClientSecret: gitHubEnterpriseSettings.OAuthClientSecret,
		}
	}

	transport := &oauth.Transport{
		Config: oauthConfig,
		Token:  &oauth.Token{AccessToken: oauthToken},
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
	gitHubClient, err := client.getGitHubClient(user.GitHubOauth)
	if err != nil {
		return nil, err
	}

	keys, _, err := gitHubClient.Users.ListKeys("")
	if err != nil {
		return nil, err
	}

	sshKeys := make([]string, 0, len(keys))
	for index, key := range keys {
		sshKeys[index] = *key.Key
	}

	return sshKeys, nil
}

func (client *GitHubClient) AddKoalitySshKeyToUser(user resources.User) error {
	gitHubClient, err := client.getGitHubClient(user.GitHubOauth)
	if err != nil {
		return err
	}

	keyPair, err := client.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		return err
	}

	keyTitle := "Koality Verification"

	_, _, err = gitHubClient.Users.CreateKey(&github.Key{
		Title: &keyTitle,
		Key:   &keyPair.PublicKey,
	})
	return err
}

func (client *GitHubClient) GetRepositories(user resources.User) ([]GitHubRepository, error) {
	gitHubClient, err := client.getGitHubClient(user.GitHubOauth)
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
