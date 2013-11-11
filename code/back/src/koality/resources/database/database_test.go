package database

import (
	"testing"
)

func TestConnecting(testing *testing.T) {
	New()
	err := New()
	if err != nil {
		testing.Error(err)
	}
}
