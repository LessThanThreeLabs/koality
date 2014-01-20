package settings

import (
	"database/sql"
	"encoding/json"
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

func (updateHandler *UpdateHandler) setSetting(resource, key string, value interface{}) error {
	doesSettingExist := func(resource, key string) (bool, error) {
		query := "SELECT value FROM settings WHERE resource=$1 AND key=$2"
		row := updateHandler.database.QueryRow(query, resource, key)
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

	settingExists, err := doesSettingExist(resource, key)
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
		query = "UPDATE settings SET value=$3 WHERE resource=$1 AND key=$2"
	} else {
		query = "INSERT INTO settings (resource, key, value) VALUES ($1, $2, $3)"
	}

	_, err = updateHandler.database.Exec(query, resource, key, encryptedValue)
	if err != nil {
		return err
	}
	return nil
}

func (updateHandler *UpdateHandler) ResetRepositoryKeyPair() (*resources.RepositoryKeyPair, error) {
	repositoryKeyPair, err := updateHandler.keyPairGenerator.GenerateRepositoryKeyPair()
	if err != nil {
		return nil, err
	}

	err = updateHandler.setSetting("Repository", "KeyPair", repositoryKeyPair)
	if err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireRepositoryKeyPairUpdatedEvent(repositoryKeyPair)
	return repositoryKeyPair, nil
}

func (updateHandler *UpdateHandler) SetS3ExporterSettings(accessKey, secretKey, bucketName string) (*resources.S3ExporterSettings, error) {
	if err := updateHandler.verifier.verifyEc2AccessKey(accessKey); err != nil {
		return nil, err
	} else if err := updateHandler.verifier.verifyEc2SecretKey(secretKey); err != nil {
		return nil, err
	} else if err := updateHandler.verifier.verifyEc2BucketName(bucketName); err != nil {
		return nil, err
	}

	s3Settings := &resources.S3ExporterSettings{accessKey, secretKey, bucketName}
	err := updateHandler.setSetting("Exporter", "S3Settings", s3Settings)
	if err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireS3ExporterSettingsUpdatedEvent(s3Settings)
	return s3Settings, nil
}
