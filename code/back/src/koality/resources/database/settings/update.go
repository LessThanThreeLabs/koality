package settings

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"io"
	"koality/resources"
)

type UpdateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	encrypter           *Encrypter
	keyPairGenerator    *resources.KeyPairGenerator
	subscriptionHandler resources.InternalSettingsSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, encrypter *Encrypter, keyPairGenerator *resources.KeyPairGenerator, subscriptionHandler resources.InternalSettingsSubscriptionHandler) (resources.SettingsUpdateHandler, error) {
	return &UpdateHandler{database, verifier, encrypter, keyPairGenerator, subscriptionHandler}, nil
}

func (updateHandler *UpdateHandler) setSetting(locator SettingLocator, value interface{}) error {
	doesSettingExist := func(locator SettingLocator) (bool, error) {
		query := "SELECT value FROM settings WHERE key=$1"
		row := updateHandler.database.QueryRow(query, locator.String())
		var tempBytes []byte
		err := row.Scan(&tempBytes)
		if err == sql.ErrNoRows {
			return false, nil
		} else if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}

	settingExists, err := doesSettingExist(locator)
	if err != nil {
		return err
	}

	jsonedValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	encryptedValue, err := updateHandler.encrypter.EncryptValue(jsonedValue)
	if err != nil {
		return err
	}

	var query string
	if settingExists {
		query = "UPDATE settings SET value=$2 WHERE key=$1"
	} else {
		query = "INSERT INTO settings (key, value) VALUES ($1, $2)"
	}

	_, err = updateHandler.database.Exec(query, locator.String(), encryptedValue)
	if err != nil {
		return err
	}
	return nil
}

func (updateHandler *UpdateHandler) SetDomainName(domainName string) (resources.DomainName, error) {
	if err := updateHandler.verifier.verifyDomainName(domainName); err != nil {
		return "", err
	}

	domainNameSetting := resources.DomainName(domainName)
	if err := updateHandler.setSetting(domainNameLocator, domainNameSetting); err != nil {
		return "", err
	}

	updateHandler.subscriptionHandler.FireDomainNameUpdatedEvent(domainNameSetting)
	return domainNameSetting, nil
}

func (updateHandler *UpdateHandler) SetAuthenticationSettings(manualAllowed, googleAllowed bool, allowedDomains []string) (*resources.AuthenticationSettings, error) {
	for _, domainName := range allowedDomains {
		if err := updateHandler.verifier.verifyDomainName(domainName); err != nil {
			return nil, err
		}
	}

	authenticationSettings := &resources.AuthenticationSettings{manualAllowed, googleAllowed, allowedDomains}

	if err := updateHandler.setSetting(authenticationSettingsLocator, authenticationSettings); err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireAuthenticationSettingsUpdatedEvent(authenticationSettings)
	return authenticationSettings, nil
}

func (updateHandler *UpdateHandler) ResetRepositoryKeyPair() (*resources.RepositoryKeyPair, error) {
	repositoryKeyPair, err := updateHandler.keyPairGenerator.GenerateRepositoryKeyPair()
	if err != nil {
		return nil, err
	}

	if err = updateHandler.setSetting(repositoryKeyPairLocator, repositoryKeyPair); err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireRepositoryKeyPairUpdatedEvent(repositoryKeyPair)
	return repositoryKeyPair, nil
}

func (updateHandler *UpdateHandler) SetS3ExporterSettings(accessKey, secretKey, bucketName string) (*resources.S3ExporterSettings, error) {
	if err := updateHandler.verifier.verifyAwsAccessKey(accessKey); err != nil {
		return nil, err
	} else if err := updateHandler.verifier.verifyAwsSecretKey(secretKey); err != nil {
		return nil, err
	} else if err := updateHandler.verifier.verifyS3BucketName(bucketName); err != nil {
		return nil, err
	}

	s3Settings := &resources.S3ExporterSettings{accessKey, secretKey, bucketName}
	err := updateHandler.setSetting(s3ExporterSettingsLocator, s3Settings)
	if err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireS3ExporterSettingsUpdatedEvent(s3Settings)
	return s3Settings, nil
}

func (updateHandler *UpdateHandler) SetHipChatSettings(authenticationToken string, rooms []string, notifyOn string) (*resources.HipChatSettings, error) {
	if err := updateHandler.verifier.verifyHipChatAuthenticationToken(authenticationToken); err != nil {
		return nil, err
	}

	hipChatSettings := &resources.HipChatSettings{authenticationToken, rooms, notifyOn}
	err := updateHandler.setSetting(hipChatSettingsLocator, hipChatSettings)
	if err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireHipChatSettingsUpdatedEvent(hipChatSettings)
	return hipChatSettings, nil
}

func (updateHandler *UpdateHandler) ResetCookieStoreKeys() (*resources.CookieStoreKeys, error) {
	cookieStoreKeys := new(resources.CookieStoreKeys)

	var authenticationBuffer bytes.Buffer
	_, err := io.CopyN(&authenticationBuffer, rand.Reader, 32)
	if err != nil {
		return nil, err
	}
	cookieStoreKeys.Authentication = authenticationBuffer.Bytes()

	var encryptionBuffer bytes.Buffer
	_, err = io.CopyN(&encryptionBuffer, rand.Reader, 32)
	if err != nil {
		return nil, err
	}
	cookieStoreKeys.Encryption = encryptionBuffer.Bytes()

	err = updateHandler.setSetting(cookieStoreKeysLocator, cookieStoreKeys)
	if err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireCookieStoreKeysUpdatedEvent(cookieStoreKeys)
	return cookieStoreKeys, nil
}

func (updateHandler *UpdateHandler) setSmtpServerSettings(hostname string, port uint16, smtpAuthSettings resources.SmtpAuthSettings) (*resources.SmtpServerSettings, error) {
	smtpServerSettings := &resources.SmtpServerSettings{hostname, port, smtpAuthSettings}
	if err := updateHandler.setSetting(smtpServerSettingsLocator, smtpServerSettings); err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireSmtpServerSettingsUpdatedEvent(smtpServerSettings)
	return smtpServerSettings, nil
}

func (updateHandler *UpdateHandler) SetSmtpAuthPlain(hostname string, port uint16, identity, username, password, host string) (*resources.SmtpServerSettings, error) {
	smtpAuthSettings := resources.SmtpAuthSettings{
		Plain: &resources.SmtpPlainAuthSettings{identity, username, password, host},
	}

	return updateHandler.setSmtpServerSettings(hostname, port, smtpAuthSettings)
}

func (updateHandler *UpdateHandler) SetSmtpAuthCramMd5(hostname string, port uint16, username, secret string) (*resources.SmtpServerSettings, error) {
	smtpAuthSettings := resources.SmtpAuthSettings{
		CramMd5: &resources.SmtpCramMd5AuthSettings{username, secret},
	}

	return updateHandler.setSmtpServerSettings(hostname, port, smtpAuthSettings)
}

func (updateHandler *UpdateHandler) SetSmtpAuthLogin(hostname string, port uint16, username, password string) (*resources.SmtpServerSettings, error) {
	smtpAuthSettings := resources.SmtpAuthSettings{
		Login: &resources.SmtpLoginAuthSettings{username, password},
	}

	return updateHandler.setSmtpServerSettings(hostname, port, smtpAuthSettings)
}

func (updateHandler *UpdateHandler) SetGitHubEnterpriseSettings(baseUri, oAuthClientId, oAuthClientSecret string) (*resources.GitHubEnterpriseSettings, error) {
	// TODO (bbland): verify these settings
	gitHubEnterpriseSettings := &resources.GitHubEnterpriseSettings{baseUri, oAuthClientId, oAuthClientSecret}
	err := updateHandler.setSetting(gitHubEnterpriseSettingsLocator, gitHubEnterpriseSettings)
	if err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireGitHubEnterpriseSettingsUpdatedEvent(gitHubEnterpriseSettings)
	return gitHubEnterpriseSettings, nil
}

func (updateHandler *UpdateHandler) ResetApiKey() (resources.ApiKey, error) {
	var apiKey resources.ApiKey

	var apiKeyBuffer bytes.Buffer
	_, err := io.CopyN(&apiKeyBuffer, rand.Reader, 15)
	if err != nil {
		return "", err
	}

	// See http://en.wikipedia.org/wiki/Base32#Crockford.27s_Base32
	crockfordsBase32Alphabet := "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	apiKey = resources.ApiKey(base32.NewEncoding(crockfordsBase32Alphabet).EncodeToString(apiKeyBuffer.Bytes()))

	err = updateHandler.setSetting(apiKeyLocator, apiKey)
	if err != nil {
		return "", err
	}

	updateHandler.subscriptionHandler.FireApiKeyUpdatedEvent(apiKey)
	return apiKey, nil
}
