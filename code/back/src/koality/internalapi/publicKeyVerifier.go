package internalapi

import (
	"koality/resources"
)

type PublicKeyVerifier struct {
	resourcesConnection *resources.Connection
}

func (publicKeyVerifier PublicKeyVerifier) GetUserIdForKey(publicKey string, userIdRes *uint64) error {
	userId, err := publicKeyVerifier.resourcesConnection.Users.Read.GetIdByPublicKey(publicKey)
	if err == nil {
		*userIdRes = userId
		return nil
	} else if _, ok := err.(resources.NoSuchUserError); ok {
		return nil
	} else {
		return err
	}
}
