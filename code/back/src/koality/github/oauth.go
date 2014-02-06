package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koality/resources"
	"net/http"
	"net/url"
)

type GitHubOAuthConnection interface {
	GetAuthorizationUrl(action string) (authorizationUrl string, err error)
	CheckValidOAuthToken(oAuthToken string) (isValid bool, err error)
}

type standardGitHubOAuthConnection struct{}

func NewStandardGitHubOAuthConnection() GitHubOAuthConnection {
	return standardGitHubOAuthConnection{}
}

func (connection standardGitHubOAuthConnection) GetAuthorizationUrl(action string) (string, error) {
	baseUrl := "https://127.0.0.1:10443"
	redirectUrl := baseUrl + "/oAuth/gitHub/token"
	queryValues := url.Values{}
	queryValues.Set("redirectUri", redirectUrl)
	queryValues.Set("action", action)
	return "https://koalitycode.com/gitHub/authenticate?" + queryValues.Encode(), nil
}

func (connection standardGitHubOAuthConnection) CheckValidOAuthToken(oAuthToken string) (bool, error) {
	oAuthCheckUrl := "https://koalitycode.com/gitHub/isValidOAuth"
	jsonBytes, err := json.Marshal(map[string]string{"token": oAuthToken})
	if err != nil {
		return false, err
	}

	checkOAuthTokenRequestBody := bytes.NewReader(jsonBytes)
	checkOAuthTokenRequest, err := http.NewRequest("GET", oAuthCheckUrl, checkOAuthTokenRequestBody)
	if err != nil {
		return false, err
	}

	httpClient := new(http.Client)
	response, err := httpClient.Do(checkOAuthTokenRequest)
	defer response.Body.Close()
	if err != nil {
		return false, err
	}
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Failed to check OAuthToken, received http status: %d (%s)", response.StatusCode, response.Status)
	}

	var isValid bool
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&isValid)
	if err != nil {
		return false, err
	}

	return isValid, nil
}

type gitHubEnterpriseOAuthConnection struct {
	resourcesConnection *resources.Connection
}

func NewGitHubEnterpriseOAuthConnection(resourcesConnection *resources.Connection) GitHubOAuthConnection {
	return &gitHubEnterpriseOAuthConnection{resourcesConnection}
}

func (connection *gitHubEnterpriseOAuthConnection) GetAuthorizationUrl(action string) (string, error) {
	gitHubEnterpriseSettings, err := connection.resourcesConnection.Settings.Read.GetGitHubEnterpriseSettings()
	if err != nil {
		return "", err
	}

	queryValues := url.Values{}
	queryValues.Set("client_id", gitHubEnterpriseSettings.OAuthClientId)
	queryValues.Set("scope", "user,repo")
	queryValues.Set("state", action)
	return gitHubEnterpriseSettings.BaseUrl + "/login/oauth/authorize?" + queryValues.Encode(), nil
}

func (connection *gitHubEnterpriseOAuthConnection) CheckValidOAuthToken(oAuthToken string) (bool, error) {
	gitHubEnterpriseSettings, err := connection.resourcesConnection.Settings.Read.GetGitHubEnterpriseSettings()
	if err != nil {
		return false, err
	}

	oAuthCheckUrl := fmt.Sprintf("%s/api/v3/applications/%s/tokens/%s", gitHubEnterpriseSettings.BaseUrl, gitHubEnterpriseSettings.OAuthClientId, oAuthToken)

	checkOAuthTokenRequest, err := http.NewRequest("GET", oAuthCheckUrl, nil)
	if err != nil {
		return false, err
	}

	checkOAuthTokenRequest.SetBasicAuth(gitHubEnterpriseSettings.OAuthClientId, gitHubEnterpriseSettings.OAuthClientSecret)

	httpClient := new(http.Client)
	response, err := httpClient.Do(checkOAuthTokenRequest)
	defer response.Body.Close()
	if err != nil {
		return false, err
	}
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Failed to check OAuthToken, received http status: %d (%s)", response.StatusCode, response.Status)
	}

	var oAuthResponse struct {
		Id *int64 `json:"id"`
	}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&oAuthResponse)
	if err != nil {
		return false, err
	}

	return oAuthResponse.Id != nil, nil
}

type compoundGitHubOAuthConnection struct {
	standard   GitHubOAuthConnection
	enterprise GitHubOAuthConnection
}

func NewCompoundGitHubOAuthConnection(resourcesConnection *resources.Connection) GitHubOAuthConnection {
	return &compoundGitHubOAuthConnection{
		standard:   NewStandardGitHubOAuthConnection(),
		enterprise: NewGitHubEnterpriseOAuthConnection(resourcesConnection),
	}
}

func (connection *compoundGitHubOAuthConnection) GetAuthorizationUrl(action string) (string, error) {
	authorizationUrl, err := connection.enterprise.GetAuthorizationUrl(action)
	if _, ok := err.(resources.NoSuchSettingError); err != nil && !ok {
		return "", err
	} else if err != nil {
		return connection.standard.GetAuthorizationUrl(action)
	} else {
		return authorizationUrl, nil
	}
}

func (connection *compoundGitHubOAuthConnection) CheckValidOAuthToken(oAuthToken string) (bool, error) {
	isValid, err := connection.enterprise.CheckValidOAuthToken(oAuthToken)
	if _, ok := err.(resources.NoSuchSettingError); err != nil && !ok {
		return false, err
	} else if err != nil {
		return connection.standard.CheckValidOAuthToken(oAuthToken)
	} else {
		return isValid, nil
	}
}
