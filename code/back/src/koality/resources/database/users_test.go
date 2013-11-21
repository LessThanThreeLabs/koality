package database

import (
	"testing"
)

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

	if len(users) != 1 {
		testing.Fatal("Returned incorrect number of users")
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

		if user.Admin != admin {
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

	keyId, err := connection.Users.Update.AddKey(1000, "alias", "public-key")
	if err != nil {
		testing.Fatal(err)
	}

	err = connection.Users.Update.RemoveKey(1000, keyId)
	if err != nil {
		testing.Fatal(err)
	}
}
