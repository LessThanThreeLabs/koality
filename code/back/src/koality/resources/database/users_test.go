package database

import (
	"koality/resources"
	"testing"
)

func TestCreateUser(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	userId, err := connection.Users.Create.Create("test-email@address.com", "First", "Last", "password-hash", "password-salt", false)
	if err != nil {
		testing.Fatal(err)
	}

	user, err := connection.Users.Read.Get(userId)
	if err != nil {
		testing.Fatal(err)
	}

	if user.Id != userId {
		testing.Fatal("user.Id mismatch")
	}

	_, err = connection.Users.Create.Create(user.Email, user.FirstName, user.LastName, user.PasswordHash, user.PasswordSalt, user.IsAdmin)
	if _, ok := err.(resources.UserAlreadyExistsError); !ok {
		testing.Fatal("Expected UserAlreadyExistsError when trying to add same user twice")
	}

	err = connection.Users.Delete.Delete(userId)
	if err != nil {
		testing.Fatal(err)
	}

	_, err = connection.Users.Read.Get(userId)
	if err == nil {
		testing.Fatal("Found a user that should have been deleted")
	}

	err = connection.Users.Delete.Delete(userId)
	if _, ok := err.(resources.NoSuchUserError); !ok {
		testing.Fatal("Expected NoSuchUserError when trying to delete same user twice")
	}
}

func TestSingleUserRead(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	user, err := connection.Users.Read.Get(1000)
	if err != nil {
		testing.Fatal(err)
	}

	if user.Id != 1000 {
		testing.Fatal("user.Id mismatch")
	}

	user, err = connection.Users.Read.GetByEmail(user.Email)
	if err != nil {
		testing.Fatal(err)
	}

	if user.Id != 1000 {
		testing.Fatal("user.Id mismatch")
	}
}

func TestUsersRead(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		testing.Fatal(err)
	}

	for _, user := range users {
		if user.Id < 1000 {
			testing.Fatal("user.Id < 1000")
		}
	}
}

func TestUsersUpdateName(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	updateAndCheckName := func(userId int64, firstName, lastName string) {
		err = connection.Users.Update.SetName(userId, firstName, lastName)
		if err != nil {
			testing.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			testing.Fatal(err)
		}

		if user.FirstName != firstName || user.LastName != lastName {
			testing.Fatal("Name not updated")
		}
	}

	updateAndCheckName(1000, "Jordan2", "Potter2")
	updateAndCheckName(1000, "Jordan", "Potter")
}

func TestUsersUpdatePassword(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	updateAndCheckPassword := func(userId int64, passwordHash, passwordSalt string) {
		err = connection.Users.Update.SetPassword(userId, passwordHash, passwordSalt)
		if err != nil {
			testing.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			testing.Fatal(err)
		}

		if user.PasswordHash != passwordHash || user.PasswordSalt != passwordSalt {
			testing.Fatal("Password not updated")
		}
	}

	updateAndCheckPassword(1000, "password-hash", "password-salt")
}

func TestUsersUpdateGitHubOauth(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	updateAndCheckGitHubOauth := func(userId int64, gitHubOauth string) {
		err = connection.Users.Update.SetGitHubOauth(userId, gitHubOauth)
		if err != nil {
			testing.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			testing.Fatal(err)
		}

		if user.GitHubOauth != gitHubOauth {
			testing.Fatal("GitHubOauth not updated")
		}
	}

	updateAndCheckGitHubOauth(1000, "github-oauth")
}

func TestUsersUpdateAdmin(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	updateAndCheckAdmin := func(userId int64, admin bool) {
		err = connection.Users.Update.SetAdmin(userId, admin)
		if err != nil {
			testing.Fatal(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			testing.Fatal(err)
		}

		if user.IsAdmin != admin {
			testing.Fatal("Admin status not updated")
		}
	}

	updateAndCheckAdmin(1000, false)
	updateAndCheckAdmin(1000, true)
}

func TestUsersSshKeys(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Fatal(err)
	}

	keyId, err := connection.Users.Update.AddKey(1000, "test-alias", "test-public-key")
	if err != nil {
		testing.Fatal(err)
	}

	_, err = connection.Users.Update.AddKey(1000, "test-alias", "test-public-key-2")
	if _, ok := err.(resources.KeyAlreadyExistsError); !ok {
		testing.Fatal("Expected KeyAlreadyExistsError when trying to add same key twice")
	}

	err = connection.Users.Update.RemoveKey(1000, keyId)
	if err != nil {
		testing.Fatal(err)
	}

	err = connection.Users.Update.RemoveKey(1000, keyId)
	if _, ok := err.(resources.NoSuchKeyError); !ok {
		testing.Fatal("Expected NoSuchKeyError when trying to delete same user twice")
	}
}
