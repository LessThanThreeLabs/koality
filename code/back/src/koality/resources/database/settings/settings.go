package settings

import (
	"database/sql"
	"koality/resources"
)

const (
	encryptionKey = "e36d0e19bbc78ca123540f7334896c71"
)

type SettingLocator string

func (locator SettingLocator) String() string {
	return string(locator)
}

var (
	domainNameLocator               = SettingLocator("DomainName")
	authenticationSettingsLocator   = SettingLocator("AuthenticationSettings")
	repositoryKeyPairLocator        = SettingLocator("RepositoryKeyPair")
	s3ExporterSettingsLocator       = SettingLocator("S3ExporterSettings")
	hipChatSettingsLocator          = SettingLocator("HipChatSettings")
	cookieStoreKeysLocator          = SettingLocator("CookieStoreKeys")
	smtpServerSettingsLocator       = SettingLocator("SmtpServerSettings")
	gitHubEnterpriseSettingsLocator = SettingLocator("GitHubEnterpriseSettings")
	apiKeyLocator                   = SettingLocator("ApiKey")
)

func New(database *sql.DB) (*resources.SettingsHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}

	encrypter, err := NewEncrypter()
	if err != nil {
		return nil, err
	}

	keyPairGenerator, err := resources.NewKeyPairGenerator()
	if err != nil {
		return nil, err
	}

	internalSubscriptionHandler, err := NewInternalSubscriptionHandler()
	if err != nil {
		return nil, err
	}

	readHandler, err := NewReadHandler(database, verifier, encrypter, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	updateHandler, err := NewUpdateHandler(database, verifier, encrypter, keyPairGenerator, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	deleteHandler, err := NewDeleteHandler(database, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	return &resources.SettingsHandler{readHandler, updateHandler, deleteHandler, internalSubscriptionHandler}, nil
}
