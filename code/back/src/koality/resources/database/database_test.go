package database

import (
	"testing"
)

func TestConnecting(testing *testing.T) {
	_, err := New()
	if err != nil {
		testing.Error(err)
	}
}

func TestUsersRead(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Error(err)
	}

	user, err := connection.Users.Read.Get(1000)
	if err != nil {
		testing.Error(err)
	}

	if user.Id != 1000 {
		testing.Error("user.Id mismatch")
	}

	user, err = connection.Users.Read.GetByEmail(user.Email)
	if err != nil {
		testing.Error(err)
	}

	if user.Id != 1000 {
		testing.Error("user.Id mismatch")
	}
}

func TestUsersUpdate(testing *testing.T) {
	connection, err := New()
	if err != nil {
		testing.Error(err)
	}

	updateAndCheckName := func(userId int, firstName, lastName string) {
		err = connection.Users.Update.SetName(userId, firstName, lastName)
		if err != nil {
			testing.Error(err)
		}

		user, err := connection.Users.Read.Get(userId)
		if err != nil {
			testing.Error(err)
		}

		if user.FirstName != firstName || user.LastName != lastName {
			testing.Error("Name not updated")
		}
	}

	updateAndCheckName(1000, "Jordan2", "Potter2")
	updateAndCheckName(1000, "Jordan", "Potter")
}
