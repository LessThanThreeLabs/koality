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

func (readHandler *ReadHandler) getSetting(resource, key string) ([]byte, error) {
	var value []byte
	query := "SELECT value FROM settings WHERE resource=$1 AND key=$2"
	row := readHandler.database.QueryRow(query, resource, key)
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find setting %s-%s", resource, key)
		return nil, resources.NoSuchSettingError{errorText}
	} else if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return nil, nil
	}
	return readHandler.encrypter.DecryptValue(value)
}

func (readHandler *ReadHandler) GetRepositoryKeyPair() (*resources.RepositoryKeyPair, error) {
	jsonedKeyPair, err := readHandler.getSetting("Repository", "KeyPair")
	if err != nil {
		return nil, err
	}

	repositoryKeyPair := new(resources.RepositoryKeyPair)
	err = json.Unmarshal(jsonedKeyPair, repositoryKeyPair)
	if err != nil {
		return nil, err
	}
	return repositoryKeyPair, nil
}
