package database

import (
	"testing"
)

func TestSettingsResetRepositoryKeyPair(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryKeyPair, err := connection.Settings.Update.ResetRepositoryKeyPair()
	if err != nil {
		test.Fatal(err)
	}

	repositoryKeyPair2, err := connection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		test.Fatal(err)
	}

	if repositoryKeyPair.PrivateKey != repositoryKeyPair2.PrivateKey {
		test.Fatal("PrivateKey mismatch")
	} else if repositoryKeyPair.PublicKey != repositoryKeyPair2.PublicKey {
		test.Fatal("PublicKey mismatch")
	}

	repositoryKeyPair3, err := connection.Settings.Update.ResetRepositoryKeyPair()
	if err != nil {
		test.Fatal(err)
	}

	if repositoryKeyPair.PrivateKey == repositoryKeyPair3.PrivateKey {
		test.Fatal("Expected new PrivateKey")
	} else if repositoryKeyPair.PublicKey == repositoryKeyPair3.PublicKey {
		test.Fatal("Expected new PublicKey")
	}
}
