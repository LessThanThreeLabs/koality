package settings

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"koality/resources"
)

type UpdateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	encrypter           *Encrypter
	subscriptionHandler resources.InternalSettingsSubscriptionHandler
}

func NewUpdateHandler(database *sql.DB, verifier *Verifier, encrypter *Encrypter, subscriptionHandler resources.InternalSettingsSubscriptionHandler) (resources.SettingsUpdateHandler, error) {
	return &UpdateHandler{database, verifier, encrypter, subscriptionHandler}, nil
}

func (updateHandler *UpdateHandler) setSetting(resource, key string, value []byte) error {
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

	encryptedValue, err := updateHandler.encrypter.EncryptValue(value)
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
	repositoryKeyPair, err := updateHandler.generateRepositoryKeyPair()
	if err != nil {
		return nil, err
	}

	jsonedKeyPair, err := json.Marshal(repositoryKeyPair)
	if err != nil {
		return nil, err
	}

	err = updateHandler.setSetting("Repository", "KeyPair", jsonedKeyPair)
	if err != nil {
		return nil, err
	}

	updateHandler.subscriptionHandler.FireRepositoryKeyPairUpdatedEvent(repositoryKeyPair)
	return repositoryKeyPair, nil
}

func (updateHandler *UpdateHandler) generateRepositoryKeyPair() (*resources.RepositoryKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, err
	}

	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	privateKeyPem := string(pem.EncodeToMemory(&privateKeyBlock))

	publicKey := privateKey.PublicKey
	publicKeyDer, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return nil, err
	}

	// generate public key with "sha-rsa " + encode64(publicKeyDer)
	//
	//
	//

	publicKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDer,
	}
	publicKeyPem := string(pem.EncodeToMemory(&publicKeyBlock))

	repositoryKeyPair := resources.RepositoryKeyPair{
		PrivateKey: privateKeyPem,
		PublicKey:  publicKeyPem,
	}
	return &repositoryKeyPair, nil
}
