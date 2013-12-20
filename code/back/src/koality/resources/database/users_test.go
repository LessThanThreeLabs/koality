package database

import (
	"bytes"
	"koality/resources"
	"testing"
)

func TestCreateInvalidUser(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	email := "test-email@address.com"
	firstName := "First"
	lastName := "Last"
	passwordHash := []byte("password-hash")
	passwordSalt := []byte("password-hash")

	_, err = connection.Users.Create.Create("", firstName, lastName, passwordHash, passwordSalt, false)
	if err == nil {
		test.Fatal("Expected error after providing invalid email")
	}

	_, err = connection.Users.Create.Create(email, "", lastName, passwordHash, passwordSalt, false)
	if err == nil {
		test.Fatal("Expected error after providing invalid first name")
	}

	_, err = connection.Users.Create.Create(email, "!!First", lastName, passwordHash, passwordSalt, false)
	if err == nil {
		test.Fatal("Expected error after providing invalid first name")
	}

	_, err = connection.Users.Create.Create(email, firstName, "1234", passwordHash, passwordSalt, false)
	if err == nil {
		test.Fatal("Expected error after providing invalid last name")
	}

	_, err = connection.Users.Create.Create(email, firstName, "Last$", passwordHash, passwordSalt, false)
	if err == nil {
		test.Fatal("Expected error after providing invalid last name")
	}
}

func TestCreateAndDeleteUser(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	userCreatedEventId := uint64(0)
	userCreatedHandler := func(userId uint64) {
		userCreatedEventId = userId
	}
	_, err = connection.Users.Subscription.SubscribeToUserCreatedEvents(userCreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	userDeletedEventId := uint64(0)
	userDeletedHandler := func(userId uint64) {
		userDeletedEventId = userId
	}
	_, err = connection.Users.Subscription.SubscribeToUserDeletedEvents(userDeletedHandler)
	if err != nil {
		test.Fatal(err)
	}

	userId, err := connection.Users.Create.Create("test-email@address.com", "First", "Last", []byte("password-hash"), []byte("password-salt"), false)
	if err != nil {
		test.Fatal(err)
	}

	if userCreatedEventId != userId {
		test.Fatal("Bad userId in creation event")
	}

	user, err := connection.Users.Read.Get(userId)
	if err != nil {
		test.Fatal(err)
	}

	if user.Id != userId {
		test.Fatal("user.Id mismatch")
	}

	_, err = connection.Users.Create.Create(user.Email, user.FirstName, user.LastName, user.PasswordHash, user.PasswordSalt, user.IsAdmin)
	if _, ok := err.(resources.UserAlreadyExistsError); !ok {
		test.Fatal("Expected UserAlreadyExistsError when trying to add same user twice")
	}

	err = connection.Users.Delete.Delete(userId)
	if err != nil {
		test.Fatal(err)
	}

	if userDeletedEventId != userId {
		test.Fatal("Bad userId in deletion event")
	}

	err = connection.Users.Delete.Delete(userId)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to delete same user twice")
	}
}

func TestUsersRead(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	if len(users) != 4 {
		test.Fatal("Expected 4 users")
	}

	for _, user := range users {
		if user.Id < 1000 {
			test.Fatal("user.Id < 1000")
		}
	}

	firstUser := users[0]
	user, err := connection.Users.Read.Get(firstUser.Id)
	if err != nil {
		test.Fatal(err)
	}

	if user.Id != firstUser.Id {
		test.Fatal("user.Id mismatch")
	}

	user, err = connection.Users.Read.GetByEmail(user.Email)
	if err != nil {
		test.Fatal(err)
	}

	if user.Id != firstUser.Id {
		test.Fatal("user.Id mismatch")
	}
}

func TestUsersUpdateName(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckName := func(userId uint64, firstName, lastName string) {
		userEventId := uint64(0)
		userEventFirstName := ""
		userEventLastName := ""
		userNameUpdatedHandler := func(userId uint64, firstName, lastName string) {
			userEventId = userId
			userEventFirstName = firstName
			userEventLastName = lastName
		}
		subscriptionId, err := connection.Users.Subscription.SubscribeToNameUpdatedEvents(userNameUpdatedHandler)
		if err != nil {
			test.Fatal(err)
		}

		err = connection.Users.Update.SetName(userId, firstName, lastName)
		if err != nil {
			test.Fatal(err)
		}

		if userEventId != userId {
			test.Fatal("Bad userId in name updated event")
		} else if userEventFirstName != firstName {
			test.Fatal("Bad firstName in name updated event")
		} else if userEventLastName != lastName {
			test.Fatal("Bad lastName in name updated event")
		}

		err = connection.Users.Subscription.UnsubscribeFromNameUpdatedEvents(subscriptionId)
		if err != nil {
			test.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			test.Fatal(err)
		}

		if user.FirstName != firstName || user.LastName != lastName {
			test.Fatal("Name not updated")
		}
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	firstUser := users[0]
	updateAndCheckName(firstUser.Id, "McJordan", "McPotter")
	updateAndCheckName(firstUser.Id, "First", "Last")

	lastUser := users[len(users)-1]
	updateAndCheckName(lastUser.Id, "First", "Last")
	updateAndCheckName(lastUser.Id, "McJordan", "McPotter")

	err = connection.Users.Update.SetName(13370, "First", "Last")
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set name for nonexistent user")
	}
}

func TestUsersUpdatePassword(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckPassword := func(userId uint64, passwordHash, passwordSalt []byte) {
		err = connection.Users.Update.SetPassword(userId, passwordHash, passwordSalt)
		if err != nil {
			test.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			test.Fatal(err)
		}

		if bytes.Compare(user.PasswordHash, passwordHash) != 0 || bytes.Compare(user.PasswordSalt, passwordSalt) != 0 {
			test.Fatal("Password not updated")
		}
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	firstUser := users[0]
	updateAndCheckPassword(firstUser.Id, []byte("password-hash-2"), []byte("password-salt-2"))
	updateAndCheckPassword(firstUser.Id, []byte("password-hash"), []byte("password-salt"))

	lastUser := users[len(users)-1]
	updateAndCheckPassword(lastUser.Id, []byte("password-hash"), []byte("password-salt"))
	updateAndCheckPassword(lastUser.Id, []byte("password-hash-2"), []byte("password-salt-2"))

	err = connection.Users.Update.SetPassword(13370, []byte("password-hash"), []byte("password-salt"))
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set password for nonexistent user")
	}
}

func TestUsersUpdateGitHubOauth(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckGitHubOauth := func(userId uint64, gitHubOauth string) {
		err = connection.Users.Update.SetGitHubOauth(userId, gitHubOauth)
		if err != nil {
			test.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			test.Fatal(err)
		}

		if user.GitHubOauth != gitHubOauth {
			test.Fatal("GitHubOauth not updated")
		}
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	firstUser := users[0]
	updateAndCheckGitHubOauth(firstUser.Id, "github-oauth")
	updateAndCheckGitHubOauth(firstUser.Id, "github-oauth-2")

	lastUser := users[len(users)-1]
	updateAndCheckGitHubOauth(lastUser.Id, "github-oauth-2")
	updateAndCheckGitHubOauth(lastUser.Id, "github-oauth")

	err = connection.Users.Update.SetGitHubOauth(13370, "github-oauth")
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set github oauth for nonexistent user")
	}
}

func TestUsersUpdateAdmin(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckAdmin := func(userId uint64, admin bool) {
		userEventId := uint64(0)
		userEventAdmin := false
		userAdminUpdatedHandler := func(userId uint64, admin bool) {
			userEventId = userId
			userEventAdmin = admin
		}
		subscriptionId, err := connection.Users.Subscription.SubscribeToAdminUpdatedEvents(userAdminUpdatedHandler)
		if err != nil {
			test.Fatal(err)
		}

		err = connection.Users.Update.SetAdmin(userId, admin)
		if err != nil {
			test.Fatal(err)
		}

		if userEventId != userId {
			test.Fatal("Bad userId in admin updated event")
		} else if userEventAdmin != admin {
			test.Fatal("Bad admin in admin updated event")
		}

		err = connection.Users.Subscription.UnsubscribeFromAdminUpdatedEvents(subscriptionId)
		if err != nil {
			test.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			test.Fatal(err)
		}

		if user.IsAdmin != admin {
			test.Fatal("Admin status not updated")
		}
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	firstUser := users[0]
	updateAndCheckAdmin(firstUser.Id, false)
	updateAndCheckAdmin(firstUser.Id, true)

	lastUser := users[len(users)-1]
	updateAndCheckAdmin(lastUser.Id, true)
	updateAndCheckAdmin(lastUser.Id, false)

	err = connection.Users.Update.SetAdmin(13370, true)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set admin status for nonexistent user")
	}
}

func TestUsersSshKeys(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	userSshKeyAddedEventUserId := uint64(0)
	userSshKeyAddedEventSshKeyId := uint64(0)
	userSshKeyAddedHandler := func(userId, sshKeyId uint64) {
		userSshKeyAddedEventUserId = userId
		userSshKeyAddedEventSshKeyId = sshKeyId
	}
	_, err = connection.Users.Subscription.SubscribeToSshKeyAddedEvents(userSshKeyAddedHandler)
	if err != nil {
		test.Fatal(err)
	}

	userSshKeyRemovedEventUserId := uint64(0)
	userSshKeyRemovedEventSshKeyId := uint64(0)
	userSshKeyRemovedHandler := func(userId, sshKeyId uint64) {
		userSshKeyRemovedEventUserId = userId
		userSshKeyRemovedEventSshKeyId = sshKeyId
	}
	_, err = connection.Users.Subscription.SubscribeToSshKeyRemovedEvents(userSshKeyRemovedHandler)
	if err != nil {
		test.Fatal(err)
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	firstUser := users[0]

	testPublicKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxvvK4FBlsGz6xbr5IME" +
		"fvp0LfaPg2LHJlrHPqawe66136PrXPQHDJUN5rUb8LEulVVMsW6fRjG5oAytmOZ/DCGlxLN7" +
		"vN65c8adw67lLjHVpQ8uHJteDkq0EuL/rZSPBLm2fP/yAeJYRiJP6fob24PpklwIz5cr9tGH" +
		"H7DJmzk69PzU3AdL7DbUZAvIay9cPFV5sQ3B2TpHSQlKunWWtN+m6Lp5ZAwY6+bvdw9E/8PY" +
		"p7+aBRpbPDJ4f3uiMzcmzSxAqcoz+PuCzljHeYmm/vYF2XmeB66cAzPSig3xAz5YVgTFBW9F" +
		"Wvg6W5DcdPsUQGqeyJta7ppIQW88HOpNk5 jordannpotter@gmail.com"
	keyId, err := connection.Users.Update.AddKey(firstUser.Id, "test-name", testPublicKey1)
	if err != nil {
		test.Fatal(err)
	}

	if userSshKeyAddedEventUserId != firstUser.Id {
		test.Fatal("Bad userId in ssh key added event")
	} else if userSshKeyAddedEventSshKeyId != keyId {
		test.Fatal("Bad sshKeyId in ssh key added event")
	}

	testPublicKey2 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxvvK4FBlsDz6xbr5IME" +
		"fvp0LfaPg2LHJlrHPqawe66136PrXPQHDJUN5rUb8LEulVVMsW6fRjG5oAytmOZ/DCGlxLN7" +
		"vN65c8adw67lLjHVpQ8uHJDeDkq0EuL/rZSPBLm2fP/yAeJYRiJP6fob24PpklwIz5cr9tGH" +
		"H7DJmzk69DzU3AdL7DbUZAvIay9cPFV5sQ3B2TpHSQlKunWWtN+m6Lp5ZAwY6+bvdw9E/8PY" +
		"p7+aBRpbPDJ4f3uiMzAmzSxAqcoz+PuCzljHeYmm/vYF2XmeB66cAzPSig3xAz5YVgTFBW9F" +
		"Wvg6W5DcdPsUQGqeyJta7ppIQW88HOpNk5 jordannpotter@gmail.com"
	_, err = connection.Users.Update.AddKey(firstUser.Id, "test-name", testPublicKey2)
	if _, ok := err.(resources.KeyAlreadyExistsError); !ok {
		test.Fatal("Expected KeyAlreadyExistsError when trying to add same key twice")
	}

	_, err = connection.Users.Update.AddKey(13370, "test-name", testPublicKey2)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to add ssh key for nonexistent user")
	}

	err = connection.Users.Update.RemoveKey(firstUser.Id, keyId)
	if err != nil {
		test.Fatal(err)
	}

	if userSshKeyRemovedEventUserId != firstUser.Id {
		test.Fatal("Bad userId in ssh key removed event")
	} else if userSshKeyRemovedEventSshKeyId != keyId {
		test.Fatal("Bad sshKeyId in ssh key removed event")
	}

	err = connection.Users.Update.RemoveKey(firstUser.Id, keyId)
	if _, ok := err.(resources.NoSuchKeyError); !ok {
		test.Fatal("Expected NoSuchKeyError when trying to delete same user twice")
	}

	err = connection.Users.Update.RemoveKey(13370, 17)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to remove ssh key for nonexistent user")
	}
}
