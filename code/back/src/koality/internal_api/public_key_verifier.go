package internal_api

import (
	"koality/resources"
)

type PublicKeyVerifier struct {
	resourcesConnection *resources.Connection
}

func (publicKeyVerifier PublicKeyVerifier) CheckKey(publicKey string, isValid *bool) error {
	userId, err := publicKeyVerifier.resourcesConnection.Users.Read.GetIdByKey(publicKey)
	if err == nil {
		*isValid = userId != nil
		return nil
	} else if _, ok := err.(resources.NoSuchUserError); ok {
		*isValid = false
		return nil
	} else {
		return err
	}
}
