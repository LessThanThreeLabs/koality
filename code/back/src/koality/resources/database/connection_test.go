package database

import (
	"testing"
)

func TestConnecting(test *testing.T) {
	_, err := New()
	if err != nil {
		test.Fatal(err)
	}
}
