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

func (readHandler *ReadHandler) getSetting(resource, key string, destination interface{}) error {
	var value []byte
	query := "SELECT value FROM settings WHERE resource=$1 AND key=$2"
	row := readHandler.database.QueryRow(query, resource, key)
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find setting %s-%s", resource, key)
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
	err := readHandler.getSetting("Repository", "KeyPair", repositoryKeyPair)
	if err != nil {
		return nil, err
	}
	return repositoryKeyPair, nil
}

func (readHandler *ReadHandler) GetS3ExporterSettings() (*resources.S3ExporterSettings, error) {
	s3Settings := new(resources.S3ExporterSettings)
	err := readHandler.getSetting("Exporter", "S3Settings", s3Settings)
	if err != nil {
		return nil, err
	}
	return s3Settings, nil
}
