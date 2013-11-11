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
}
