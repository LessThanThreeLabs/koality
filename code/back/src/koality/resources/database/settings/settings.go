package settings

import (
	"database/sql"
	"koality/resources"
)

const (
	encryptionKey = "e36d0e19bbc78ca123540f7334896c71"
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

	internalSubscriptionHandler, err := NewInternalSubscriptionHandler()
	if err != nil {
		return nil, err
	}

	readHandler, err := NewReadHandler(database, verifier, encrypter, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	updateHandler, err := NewUpdateHandler(database, verifier, encrypter, internalSubscriptionHandler)
	if err != nil {
		return nil, err
	}

	return &resources.SettingsHandler{readHandler, updateHandler, internalSubscriptionHandler}, nil
}
