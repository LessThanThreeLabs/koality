package users_test

import (
	"bytes"
	"koality/resources"
	"koality/resources/database"
	"testing"
	"time"
)

func TestCreateInvalidUser(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

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
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	createdEventReceived := make(chan bool, 1)
	var createdEventUser *resources.User
	userCreatedHandler := func(user *resources.User) {
		createdEventUser = user
		createdEventReceived <- true
	}
	_, err = connection.Users.Subscription.SubscribeToCreatedEvents(userCreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	deletedEventReceived := make(chan bool, 1)
	deletedEventId := uint64(0)
	userDeletedHandler := func(userId uint64) {
		deletedEventId = userId
		deletedEventReceived <- true
	}
	_, err = connection.Users.Subscription.SubscribeToDeletedEvents(userDeletedHandler)
	if err != nil {
		test.Fatal(err)
	}

	email := "test-email@address.com"
	firstName := "First"
	lastName := "Last"
	passwordHash := []byte("password-hash")
	passwordSalt := []byte("password-salt")
	isAdmin := false
	user, err := connection.Users.Create.Create(email, firstName, lastName, passwordHash, passwordSalt, isAdmin)
	if err != nil {
		test.Fatal(err)
	}

	if user.Email != email {
		test.Fatal("user.Email mismatch")
	} else if user.FirstName != firstName {
		test.Fatal("user.FirstName mismatch")
	} else if user.LastName != lastName {
		test.Fatal("user.LastName mismatch")
	} else if bytes.Compare(user.PasswordHash, passwordHash) != 0 {
		test.Fatal("user.PasswordHash mismatch")
	} else if bytes.Compare(user.PasswordSalt, passwordSalt) != 0 {
		test.Fatal("user.PasswordSalt mismatch")
	} else if user.IsAdmin != isAdmin {
		test.Fatal("user.IsAdmin mismatch")
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear user creation event")
	}

	if createdEventUser.Id != user.Id {
		test.Fatal("Bad user.Id in user creation event")
	} else if createdEventUser.Email != user.Email {
		test.Fatal("Bad user.Email in user creation event")
	} else if createdEventUser.FirstName != user.FirstName {
		test.Fatal("Bad user.FirstName in user creation event")
	} else if createdEventUser.LastName != user.LastName {
		test.Fatal("Bad user.LastName in user creation event")
	} else if bytes.Compare(createdEventUser.PasswordHash, user.PasswordHash) != 0 {
		test.Fatal("Bad user.PasswordHash in user creation event")
	} else if bytes.Compare(createdEventUser.PasswordSalt, user.PasswordSalt) != 0 {
		test.Fatal("Bad user.PasswordSalt in user creation event")
	} else if createdEventUser.IsAdmin != user.IsAdmin {
		test.Fatal("Bad user.IsAdmin in user creation event")
	}

	user2, err := connection.Users.Read.Get(user.Id)
	if err != nil {
		test.Fatal(err)
	}

	if user.Id != user2.Id {
		test.Fatal("user.Id mismatch")
	} else if user.Email != user2.Email {
		test.Fatal("user.Email mismatch")
	} else if user.FirstName != user2.FirstName {
		test.Fatal("user.FirstName mismatch")
	} else if user.LastName != user2.LastName {
		test.Fatal("user.LastName mismatch")
	} else if bytes.Compare(user.PasswordHash, user2.PasswordHash) != 0 {
		test.Fatal("user.PasswordHash mismatch")
	} else if bytes.Compare(user.PasswordSalt, user2.PasswordSalt) != 0 {
		test.Fatal("user.PasswordSalt mismatch")
	} else if user.IsAdmin != user2.IsAdmin {
		test.Fatal("user.IsAdmin mismatch")
	}

	_, err = connection.Users.Create.Create(user.Email, user.FirstName, user.LastName, user.PasswordHash, user.PasswordSalt, user.IsAdmin)
	if _, ok := err.(resources.UserAlreadyExistsError); !ok {
		test.Fatal("Expected UserAlreadyExistsError when trying to add same user twice")
	}

	err = connection.Users.Delete.Delete(user.Id)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-deletedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear user deletion event")
	}

	if deletedEventId != user.Id {
		test.Fatal("Bad user.Id in user deletion event")
	}

	err = connection.Users.Delete.Delete(user.Id)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to delete same user twice")
	}

	deletedUser, err := connection.Users.Read.Get(user.Id)
	if err != nil {
		test.Fatal(err)
	} else if !deletedUser.IsDeleted {
		test.Fatal("Expected user to be marked as deleted")
	}
}

func TestUsersRead(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

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

	_, err = connection.Users.Update.AddKey(user.Id, "akey", "ssh-rsa abc")
	if err != nil {
		test.Fatal(err)
	}

	userId, err := connection.Users.Read.GetIdByPublicKey("ssh-rsa abc")
	if err != nil {
		test.Fatal(err)
	}

	if userId != user.Id {
		test.Fatal("expected userId to equal user.Id")
	}

}

func TestUsersUpdateName(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	updateAndCheckName := func(userId uint64, firstName, lastName string) {
		userEventReceived := make(chan bool, 1)
		userEventId := uint64(0)
		userEventFirstName := ""
		userEventLastName := ""
		userNameUpdatedHandler := func(userId uint64, firstName, lastName string) {
			userEventId = userId
			userEventFirstName = firstName
			userEventLastName = lastName
			userEventReceived <- true
		}
		subscriptionId, err := connection.Users.Subscription.SubscribeToNameUpdatedEvents(userNameUpdatedHandler)
		if err != nil {
			test.Fatal(err)
		}

		err = connection.Users.Update.SetName(userId, firstName, lastName)
		if err != nil {
			test.Fatal(err)
		}

		select {
		case <-userEventReceived:
		case <-time.After(10 * time.Second):
			test.Fatal("Failed to hear user name updated event")
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

	err = connection.Users.Update.SetName(0, "First", "Last")
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set name for nonexistent user")
	}
}

func TestUsersUpdatePassword(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

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

	err = connection.Users.Update.SetPassword(0, []byte("password-hash"), []byte("password-salt"))
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set password for nonexistent user")
	}
}

func TestUsersUpdateGitHubOAuth(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	updateAndCheckGitHubOAuth := func(userId uint64, gitHubOAuth string) {
		err = connection.Users.Update.SetGitHubOAuth(userId, gitHubOAuth)
		if err != nil {
			test.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			test.Fatal(err)
		}

		if user.GitHubOAuth != gitHubOAuth {
			test.Fatal("GitHubOAuth not updated")
		}
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	firstUser := users[0]
	updateAndCheckGitHubOAuth(firstUser.Id, "github-oauth")
	updateAndCheckGitHubOAuth(firstUser.Id, "github-oauth-2")

	lastUser := users[len(users)-1]
	updateAndCheckGitHubOAuth(lastUser.Id, "github-oauth-2")
	updateAndCheckGitHubOAuth(lastUser.Id, "github-oauth")

	err = connection.Users.Update.SetGitHubOAuth(0, "github-oauth")
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set github oauth for nonexistent user")
	}
}

func TestUsersUpdateAdmin(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	updateAndCheckAdmin := func(userId uint64, admin bool) {
		userEventReceived := make(chan bool, 1)
		userEventId := uint64(0)
		userEventAdmin := false
		userAdminUpdatedHandler := func(userId uint64, admin bool) {
			userEventId = userId
			userEventAdmin = admin
			userEventReceived <- true
		}
		subscriptionId, err := connection.Users.Subscription.SubscribeToAdminUpdatedEvents(userAdminUpdatedHandler)
		if err != nil {
			test.Fatal(err)
		}

		err = connection.Users.Update.SetAdmin(userId, admin)
		if err != nil {
			test.Fatal(err)
		}

		select {
		case <-userEventReceived:
		case <-time.After(10 * time.Second):
			test.Fatal("Failed to hear user admin updated event")
		}

		if userEventId != userId {
			test.Fatal("Bad userId in user admin updated event")
		} else if userEventAdmin != admin {
			test.Fatal("Bad admin in user admin updated event")
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

	err = connection.Users.Update.SetAdmin(0, true)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to set admin status for nonexistent user")
	}
}

func TestUsersSshKeys(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	userSshKeyAddedEventReceived := make(chan bool, 1)
	userSshKeyAddedEventUserId := uint64(0)
	userSshKeyAddedEventSshKeyId := uint64(0)
	userSshKeyAddedHandler := func(userId, sshKeyId uint64) {
		userSshKeyAddedEventUserId = userId
		userSshKeyAddedEventSshKeyId = sshKeyId
		userSshKeyAddedEventReceived <- true
	}
	_, err = connection.Users.Subscription.SubscribeToSshKeyAddedEvents(userSshKeyAddedHandler)
	if err != nil {
		test.Fatal(err)
	}

	userSshKeyRemovedEventReceived := make(chan bool, 1)
	userSshKeyRemovedEventUserId := uint64(0)
	userSshKeyRemovedEventSshKeyId := uint64(0)
	userSshKeyRemovedHandler := func(userId, sshKeyId uint64) {
		userSshKeyRemovedEventUserId = userId
		userSshKeyRemovedEventSshKeyId = sshKeyId
		userSshKeyRemovedEventReceived <- true
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

	select {
	case <-userSshKeyAddedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear user ssh key added event")
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

	_, err = connection.Users.Update.AddKey(0, "test-name", testPublicKey2)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to add ssh key for nonexistent user")
	}

	err = connection.Users.Update.RemoveKey(firstUser.Id, keyId)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-userSshKeyRemovedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear user ssh key removed event")
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

	err = connection.Users.Update.RemoveKey(0, 17)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to remove ssh key for nonexistent user")
	}
}
