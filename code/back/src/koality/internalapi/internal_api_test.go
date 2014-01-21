package internalapi

import (
	"koality/resources/database"
	"net/rpc"
	"os"
	"testing"
	"time"
)

func TestSetup(testing *testing.T) {
	rpcSocket := "/tmp/koality-test-rpc.sock"
	if err := database.PopulateDatabase(); err != nil {
		testing.Fatal(err)
	}

	resourcesConnection, err := database.New()
	if err != nil {
		testing.Fatal(err)
	}

	go func() {
		if err := Setup(resourcesConnection, rpcSocket); err != nil {
			testing.Fatal(err)
		}
	}()

	// REVIEW(dhuang) is there a better way to do this?
	socketOpen := false
	for i := 0; i < 42 && !socketOpen; i++ {
		_, err = os.Stat(rpcSocket)
		if err == nil {
			socketOpen = true
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !socketOpen {
		testing.Fatal("socket took too long to exist")
	}

	client, err := rpc.Dial("unix", rpcSocket)
	if err != nil {
		testing.Fatal(err)
	}

	publicKey := "wrongkey"
	var userId uint64
	if err = client.Call("PublicKeyVerifier.GetUserIdForKey", &publicKey, &userId); err != nil {
		testing.Fatal(err)
	} else if userId != 0 {
		testing.Fatal("expected key to be invalid")
	}

	users, err := resourcesConnection.Users.Read.GetAll()
	if err != nil {
		testing.Fatal(err)
	} else if len(users) == 0 {
		testing.Fatal("expected nonzero number of users")
	}

	publicKey = "ssh-rsa abc"
	if _, err = resourcesConnection.Users.Update.AddKey(users[0].Id, "mykey", publicKey); err != nil {
		testing.Fatal(err)
	} else if err = client.Call("PublicKeyVerifier.GetUserIdForKey", &publicKey, &userId); err != nil {
		testing.Fatal(err)
	} else if userId == 0 {
		testing.Fatal("expected key to be valid")
	}
}
