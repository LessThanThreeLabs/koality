package settings

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"koality/resources"
	"os/exec"
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
	generatePrivatePem := func(privateKey *rsa.PrivateKey) (string, error) {
		privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyBlock := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   privateKeyDer,
		}
		return string(pem.EncodeToMemory(&privateKeyBlock)), nil
	}

	generatePublicPem := func(publicKey *rsa.PublicKey) (string, error) {
		publicKeyDer, err := x509.MarshalPKIXPublicKey(publicKey)
		if err != nil {
			return "", err
		}

		publicKeyBlock := pem.Block{
			Type:    "PUBLIC KEY",
			Headers: nil,
			Bytes:   publicKeyDer,
		}
		return string(pem.EncodeToMemory(&publicKeyBlock)), nil
	}

	generatePublicSshKey := func(publicPem string) (string, error) {
		file, err := ioutil.TempFile("", "publicKey.pem")
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = file.Write([]byte(publicPem))
		if err != nil {
			return "", err
		}

		command := exec.Command("ssh-keygen", "-m", "PKCS8", "-f", file.Name(), "-i")
		publicSshKey, err := command.Output()
		if err != nil {
			return "", err
		}

		return string(publicSshKey), nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, err
	}

	privateKeyPem, err := generatePrivatePem(privateKey)
	if err != nil {
		return nil, err
	}

	publicKeyPem, err := generatePublicPem(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	publicSshKey, err := generatePublicSshKey(publicKeyPem)
	if err != nil {
		return nil, err
	}

	repositoryKeyPair := resources.RepositoryKeyPair{
		PrivateKey: privateKeyPem,
		PublicKey:  publicSshKey,
	}
	return &repositoryKeyPair, nil
}
