package database

import (
	"koality/resources"
	"testing"
)

const (
	userEmail        = "test-email@address.com"
	userFirstName    = "First"
	userLastName     = "Last"
	userPasswordHash = "password-hash"
	userPasswordSalt = "password-salt"
	userAdmin        = false
)

func TestCreateUser(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	userId, err := connection.Users.Create.Create(userEmail, userFirstName, userLastName, userPasswordHash, userPasswordSalt, userAdmin)
	if err != nil {
		test.Fatal(err)
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

	_, err = connection.Users.Read.Get(userId)
	if err == nil {
		test.Fatal("Found a user that should have been deleted")
	}

	err = connection.Users.Delete.Delete(userId)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		test.Fatal("Expected NoSuchUserError when trying to delete same user twice")
	}
}

// TODO: IS EXPECTING SOME STATE ALREADY IN DATABASE
func TestSingleUserRead(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	user, err := connection.Users.Read.Get(1000)
	if err != nil {
		test.Fatal(err)
	}

	if user.Id != 1000 {
		test.Fatal("user.Id mismatch")
	}

	user, err = connection.Users.Read.GetByEmail(user.Email)
	if err != nil {
		test.Fatal(err)
	}

	if user.Id != 1000 {
		test.Fatal("user.Id mismatch")
	}
}

func TestUsersRead(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	for _, user := range users {
		if user.Id < 1000 {
			test.Fatal("user.Id < 1000")
		}
	}
}

// TODO: IS EXPECTING SOME STATE ALREADY IN DATABASE
func TestUsersUpdateName(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckName := func(userId int64, firstName, lastName string) {
		err = connection.Users.Update.SetName(userId, firstName, lastName)
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

	updateAndCheckName(1000, "Jordan2", "Potter2")
	updateAndCheckName(1000, "Jordan", "Potter")
}

// TODO: IS EXPECTING SOME STATE ALREADY IN DATABASE
func TestUsersUpdatePassword(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckPassword := func(userId int64, passwordHash, passwordSalt string) {
		err = connection.Users.Update.SetPassword(userId, passwordHash, passwordSalt)
		if err != nil {
			test.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			test.Fatal(err)
		}

		if user.PasswordHash != passwordHash || user.PasswordSalt != passwordSalt {
			test.Fatal("Password not updated")
		}
	}

	updateAndCheckPassword(1000, "password-hash", "password-salt")
}

// TODO: IS EXPECTING SOME STATE ALREADY IN DATABASE
func TestUsersUpdateGitHubOauth(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckGitHubOauth := func(userId int64, gitHubOauth string) {
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

	updateAndCheckGitHubOauth(1000, "github-oauth")
}

// TODO: IS EXPECTING SOME STATE ALREADY IN DATABASE
func TestUsersUpdateAdmin(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	updateAndCheckAdmin := func(userId int64, admin bool) {
		err = connection.Users.Update.SetAdmin(userId, admin)
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

	updateAndCheckAdmin(1000, false)
	updateAndCheckAdmin(1000, true)
}

// TODO: IS EXPECTING SOME STATE ALREADY IN DATABASE
func TestUsersSshKeys(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	keyId, err := connection.Users.Update.AddKey(1000, "test-alias", "test-public-key")
	if err != nil {
		test.Fatal(err)
	}

	_, err = connection.Users.Update.AddKey(1000, "test-alias", "test-public-key-2")
	if _, ok := err.(resources.KeyAlreadyExistsError); !ok {
		test.Fatal("Expected KeyAlreadyExistsError when trying to add same key twice")
	}

	err = connection.Users.Update.RemoveKey(1000, keyId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Users.Update.RemoveKey(1000, keyId)
	if _, ok := err.(resources.NoSuchKeyError); !ok {
		test.Fatal("Expected NoSuchKeyError when trying to delete same user twice")
	}
}
