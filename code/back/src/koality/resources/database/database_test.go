package database

import (
	"testing"
)

func TestConnecting(testing *testing.T) {
	err := New()
	if err != nil {
		testing.Error(err)
	}
}
