package database

import (
	"github.com/LessThanThreeLabs/gocheck"
	"koality/resources"
	"math/rand"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type DebugInstancesSuite struct {
	resourcesConnection *resources.Connection
	pools               []resources.Ec2Pool
	firstPool           resources.Ec2Pool
}

var _ = gocheck.Suite(&DebugInstancesSuite{})

func (suite *DebugInstancesSuite) SetUpTest(check *gocheck.C) {
	err := PopulateDatabase()
	check.Assert(err, gocheck.IsNil)

	suite.resourcesConnection, err = New()
	check.Assert(err, gocheck.IsNil)

	suite.pools, err = suite.resourcesConnection.Pools.Read.GetAllEc2Pools()
	check.Assert(err, gocheck.IsNil)

	suite.firstPool = suite.pools[0]
}

func (suite *DebugInstancesSuite) TearDownTest(check *gocheck.C) {
	if suite.resourcesConnection != nil {
		suite.resourcesConnection.Close()
	}
}

func getVerificationInfo() *resources.CoreVerificationInformation {
	createSha := func() string {
		shaChars := "0123456789ABCDEF"
		sha := ""
		for index := 0; index < 40; index++ {
			randomChar := shaChars[rand.Intn(len(shaChars))]
			sha = sha + string(randomChar)
		}
		return sha
	}
	return &resources.CoreVerificationInformation{
		2, createSha(), createSha(), "headmessage", "headuser", "heademail@foo.com", "foo@foo.foo",
	}
}

func (suite *DebugInstancesSuite) TestCreateGet(check *gocheck.C) {
	createdEventChan := make(chan *resources.DebugInstance, 1)
	debugInstanceCreatedHandler := func(debugInstance *resources.DebugInstance) {
		createdEventChan <- debugInstance
	}
	_, err := suite.resourcesConnection.DebugInstances.Subscription.SubscribeToCreatedEvents(debugInstanceCreatedHandler)
	check.Assert(err, gocheck.IsNil)

	instanceId := "instanceIdentifier"
	expireTime := time.Now().Add(1 * time.Minute)
	buildInfo := getVerificationInfo()
	debugInstance, err := suite.resourcesConnection.DebugInstances.Create.Create(
		suite.firstPool.Id, instanceId, &expireTime, buildInfo)
	check.Assert(err, gocheck.IsNil)
	check.Assert(debugInstance, gocheck.Not(gocheck.IsNil))

	select {
	case createdDebugInstance := <-createdEventChan:
		check.Assert(createdDebugInstance, gocheck.DeepEquals, debugInstance)
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear debug instance creation event")
	}

	debugInstance2, err := suite.resourcesConnection.DebugInstances.Read.Get(debugInstance.Id)
	check.Assert(err, gocheck.IsNil)
	check.Assert(debugInstance2, gocheck.DeepEquals, debugInstance)
}

func (suite *DebugInstancesSuite) TestGetAllRunning(check *gocheck.C) {
	instanceId := "instanceIdentifier"
	expireTime := time.Now().Add(1 * time.Minute)
	buildInfo := getVerificationInfo()
	debugInstance, err := suite.resourcesConnection.DebugInstances.Create.Create(
		suite.firstPool.Id, instanceId, &expireTime, buildInfo)
	check.Assert(err, gocheck.IsNil)
	check.Assert(debugInstance, gocheck.Not(gocheck.IsNil))

	debugInstances, err := suite.resourcesConnection.DebugInstances.Read.GetAllRunning()
	check.Assert(err, gocheck.IsNil)
	check.Assert(debugInstances, gocheck.DeepEquals, []resources.DebugInstance{*debugInstance})

	err = suite.resourcesConnection.Verifications.Update.SetStartTime(debugInstance.VerificationId, time.Now())
	check.Assert(err, gocheck.IsNil)

	err = suite.resourcesConnection.Verifications.Update.SetEndTime(debugInstance.VerificationId, time.Now())
	check.Assert(err, gocheck.IsNil)

	debugInstances, err = suite.resourcesConnection.DebugInstances.Read.GetAllRunning()
	check.Assert(err, gocheck.IsNil)
	check.Assert(debugInstances, gocheck.DeepEquals, []resources.DebugInstance{})
}
