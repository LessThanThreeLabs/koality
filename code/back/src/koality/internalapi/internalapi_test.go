package internalapi

import (
	"github.com/LessThanThreeLabs/gocheck"
	"koality/resources"
	"koality/resources/database"
	"net/rpc"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

const (
	rpcSocket = "/tmp/koality-test-rpc.sock"
)

type InternalAPISuite struct {
	resourcesConnection *resources.Connection
	client              *rpc.Client
}

var _ = gocheck.Suite(&InternalAPISuite{})

func (suite *InternalAPISuite) SetUpTest(check *gocheck.C) {
	err := database.PopulateDatabase()
	check.Assert(err, gocheck.IsNil)

	suite.resourcesConnection, err = database.New()
	check.Assert(err, gocheck.IsNil)

	err = Start(suite.resourcesConnection, rpcSocket)
	check.Assert(err, gocheck.IsNil)

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
		check.Fatalf("socket took too long to exist")
	}

	suite.client, err = rpc.Dial("unix", rpcSocket)
}
func (suite *InternalAPISuite) TearDownTest(check *gocheck.C) {
	suite.resourcesConnection.Close()

	if suite.client != nil {
		suite.client.Close()
	}

	os.Remove(rpcSocket)
}

func (suite *InternalAPISuite) TestSetup(check *gocheck.C) {
	publicKey := "wrongkey"
	var userId uint64
	err := suite.client.Call("PublicKeyVerifier.GetUserIdForKey", &publicKey, &userId)
	check.Assert(err, gocheck.IsNil)
	check.Assert(userId, gocheck.Equals, uint64(0))

	users, err := suite.resourcesConnection.Users.Read.GetAll()
	check.Assert(err, gocheck.IsNil)
	check.Assert(len(users), gocheck.Not(gocheck.Equals), 0)

	publicKey = "ssh-rsa abc"
	_, err = suite.resourcesConnection.Users.Update.AddKey(users[0].Id, "mykey", publicKey)
	check.Assert(err, gocheck.IsNil)
	err = suite.client.Call("PublicKeyVerifier.GetUserIdForKey", &publicKey, &userId)
	check.Assert(err, gocheck.IsNil)
	check.Assert(userId, gocheck.Not(gocheck.Equals), 0)
}
