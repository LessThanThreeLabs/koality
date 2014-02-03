package settings

import (
	"database/sql"
	"koality/resources"
)

const (
	encryptionKey = "e36d0e19bbc78ca123540f7334896c71"
)

type SettingLocator struct {
	Resource string
	Key      string
}

var (
	repositoryKeyPairLocator  SettingLocator = SettingLocator{"Repository", "KeyPair"}
	s3ExporterSettingsLocator SettingLocator = SettingLocator{"Exporter", "S3Settings"}
	cookieStoreKeysLocator    SettingLocator = SettingLocator{"CookieStore", "Keys"}
	smtpServerSettingsLocator SettingLocator = SettingLocator{"Smtp", "ServerSettings"}
	apiKeyLocator             SettingLocator = SettingLocator{"Api", "Key"}
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
