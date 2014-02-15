package settings

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"koality/resources"
)

type ReadHandler struct {
	database            *sql.DB
	verifier            *Verifier
	encrypter           *Encrypter
	subscriptionHandler resources.InternalSettingsSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, encrypter *Encrypter, subscriptionHandler resources.InternalSettingsSubscriptionHandler) (resources.SettingsReadHandler, error) {
	return &ReadHandler{database, verifier, encrypter, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) getSetting(locator SettingLocator, destination interface{}) error {
	var value []byte
	query := "SELECT value FROM settings WHERE key=$1"
	row := readHandler.database.QueryRow(query, locator.String())
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find setting with locator %v", locator)
		return resources.NoSuchSettingError{errorText}
	} else if err != nil {
		return err
	}

	decryptedValue, err := readHandler.encrypter.DecryptValue(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(decryptedValue, destination)
}

func (readHandler *ReadHandler) GetDomainName() (resources.DomainName, error) {
	var domainName resources.DomainName
	if err := readHandler.getSetting(domainNameLocator, &domainName); err != nil {
		return "", err
	}
	return domainName, nil
}

func (readHandler *ReadHandler) GetAuthenticationSettings() (*resources.AuthenticationSettings, error) {
	authenticationSettings := new(resources.AuthenticationSettings)
	if err := readHandler.getSetting(authenticationSettingsLocator, authenticationSettings); err != nil {
		return nil, err
	}
	return authenticationSettings, nil
}

func (readHandler *ReadHandler) GetRepositoryKeyPair() (*resources.RepositoryKeyPair, error) {
	repositoryKeyPair := new(resources.RepositoryKeyPair)
	if err := readHandler.getSetting(repositoryKeyPairLocator, repositoryKeyPair); err != nil {
		return nil, err
	}
	return repositoryKeyPair, nil
}

func (readHandler *ReadHandler) GetS3ExporterSettings() (*resources.S3ExporterSettings, error) {
	s3Settings := new(resources.S3ExporterSettings)
	if err := readHandler.getSetting(s3ExporterSettingsLocator, s3Settings); err != nil {
		return nil, err
	}
	return s3Settings, nil
}

func (readHandler *ReadHandler) GetHipChatSettings() (*resources.HipChatSettings, error) {
	hipChatSettings := new(resources.HipChatSettings)
	if err := readHandler.getSetting(hipChatSettingsLocator, hipChatSettings); err != nil {
		return nil, err
	}
	return hipChatSettings, nil
}

func (readHandler *ReadHandler) GetCookieStoreKeys() (*resources.CookieStoreKeys, error) {
	storeKeys := new(resources.CookieStoreKeys)
	if err := readHandler.getSetting(cookieStoreKeysLocator, storeKeys); err != nil {
		return nil, err
	}
	return storeKeys, nil
}

func (readHandler *ReadHandler) GetSmtpServerSettings() (*resources.SmtpServerSettings, error) {
	smtpServer := new(resources.SmtpServerSettings)
	if err := readHandler.getSetting(smtpServerSettingsLocator, smtpServer); err != nil {
		return nil, err
	}
	return smtpServer, nil
}

func (readHandler *ReadHandler) GetGitHubEnterpriseSettings() (*resources.GitHubEnterpriseSettings, error) {
	gitHubEnterpriseSettings := new(resources.GitHubEnterpriseSettings)
	if err := readHandler.getSetting(gitHubEnterpriseSettingsLocator, gitHubEnterpriseSettings); err != nil {
		return nil, err
	}
	return gitHubEnterpriseSettings, nil
}

func (readHandler *ReadHandler) GetLicenseSettings() (*resources.LicenseSettings, error) {
	licenseSettings := new(resources.LicenseSettings)
	if err := readHandler.getSetting(licenseSettingsLocator, &licenseSettings); err != nil {
		return nil, err
	}
	return licenseSettings, nil
}

func (readHandler *ReadHandler) GetApiKey() (resources.ApiKey, error) {
	var apiKey resources.ApiKey
	if err := readHandler.getSetting(apiKeyLocator, &apiKey); err != nil {
		return "", err
	}
	return apiKey, nil
}
