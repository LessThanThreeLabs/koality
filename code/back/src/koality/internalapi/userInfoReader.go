package internalapi

import (
	"koality/resources"
)

type UserInfoReader struct {
	resourcesConnection *resources.Connection
}

func (userInfoReader UserInfoReader) GetUser(userId uint64, userRes *resources.User) error {
	user, err := userInfoReader.resourcesConnection.Users.Read.Get(userId)
	if err == nil {
		*userRes = *user
		return nil
	} else if _, ok := err.(resources.NoSuchUserError); ok {
		return nil
	} else {
		return err
	}
}

func (userInfoReader UserInfoReader) GetRepoPrivateKey(_ interface{}, privateKeyRes *string) error {
	repoKeyPair, err := userInfoReader.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
        if err != nil {
          return err
        }

	*privateKeyRes = repoKeyPair.PrivateKey
	return nil
}
