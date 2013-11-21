package database

import (
	"testing"
)

func TestConnecting(testing *testing.T) {
	_, err := New()
	if err != nil {
		testing.Fatal(err)
	}
}
