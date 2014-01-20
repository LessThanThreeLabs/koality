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
	query := "SELECT value FROM settings WHERE resource=$1 AND key=$2"
	row := readHandler.database.QueryRow(query, locator.Resource, locator.Key)
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

func (readHandler *ReadHandler) GetRepositoryKeyPair() (*resources.RepositoryKeyPair, error) {
	repositoryKeyPair := new(resources.RepositoryKeyPair)
	err := readHandler.getSetting(repositoryKeyPairLocator, repositoryKeyPair)
	if err != nil {
		return nil, err
	}
	return repositoryKeyPair, nil
}

func (readHandler *ReadHandler) GetS3ExporterSettings() (*resources.S3ExporterSettings, error) {
	s3Settings := new(resources.S3ExporterSettings)
	err := readHandler.getSetting(s3ExporterSettingsLocator, s3Settings)
	if err != nil {
		return nil, err
	}
	return s3Settings, nil
}

func (readHandler *ReadHandler) GetCookieStoreKeys() (*resources.CookieStoreKeys, error) {
	storeKeys := new(resources.CookieStoreKeys)
	err := readHandler.getSetting(cookieStoreKeysLocator, storeKeys)
	if err != nil {
		return nil, err
	}
	return storeKeys, nil
}
