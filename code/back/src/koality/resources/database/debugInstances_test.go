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
	// virtualMachine      vm.VirtualMachine
	repository   *resources.Repository
	verification *resources.Verification
}

var _ = gocheck.Suite(&DebugInstancesSuite{})

func (suite *DebugInstancesSuite) SetUpSuite(check *gocheck.C) {
	err := PopulateDatabase()
	check.Assert(err, gocheck.IsNil)

	suite.resourcesConnection, err = New()
	check.Assert(err, gocheck.IsNil)

	suite.pools, err = suite.resourcesConnection.Pools.Read.GetAllEc2Pools()
	check.Assert(err, gocheck.IsNil)

	suite.firstPool = suite.pools[0]
}

func (suite *DebugInstancesSuite) TearDownSuite(check *gocheck.C) {
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
		0, createSha(), createSha(), "headmessage", "headuser", "heademail@foo.com", "foo@foo.foo",
	}
}

func (suite *DebugInstancesSuite) TestCreateDelete(check *gocheck.C) {
	createdEventChan := make(chan *resources.DebugInstance, 1)
	debugInstanceCreatedHandler := func(debugInstance *resources.DebugInstance) {
		createdEventChan <- debugInstance
	}
	_, err := suite.resourcesConnection.DebugInstances.Subscription.SubscribeToCreatedEvents(debugInstanceCreatedHandler)
	check.Assert(err, gocheck.IsNil)
	deletedEventChan := make(chan uint64, 1)
	debugInstanceDeletedHandler := func(debugInstanceId uint64) {
		deletedEventChan <- debugInstanceId
	}
	_, err = suite.resourcesConnection.DebugInstances.Subscription.SubscribeToDeletedEvents(debugInstanceDeletedHandler)
	check.Assert(err, gocheck.IsNil)

	instanceId := "instanceIdentifier"
	expireTime := time.Now().Add(1 * time.Minute)
	verificationInfo := getVerificationInfo()
	debugInstance, err := suite.resourcesConnection.DebugInstances.Create.Create(
		suite.firstPool.Id, instanceId, &expireTime, verificationInfo)

	select {
	case createdDebugInstance := <-createdEventChan:
		check.Assert(createdDebugInstance, gocheck.Equals, debugInstance)
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear debug instance creation event")
	}

	debugInstance2, err := suite.resourcesConnection.DebugInstances.Read.Get(debugInstance.Id)
	check.Assert(err, gocheck.IsNil)
	check.Assert(debugInstance2, gocheck.Equals, debugInstance)

	debugInstances, err := suite.resourcesConnection.DebugInstances.Read.GetAllRunning()
	check.Assert(err, gocheck.IsNil)
	check.Assert(debugInstances, gocheck.DeepEquals, []resources.DebugInstance{*debugInstance})

	err = suite.resourcesConnection.DebugInstances.Delete.Delete(debugInstance.Id)
	check.Assert(err, gocheck.IsNil)
	select {
	case debugInstanceId2 := <-deletedEventChan:
		check.Assert(debugInstanceId2, gocheck.Equals, debugInstance.Id)
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear debug instance creation event")
	}
	err = suite.resourcesConnection.DebugInstances.Delete.Delete(debugInstance.Id)
	check.Assert(err, gocheck.Not(gocheck.IsNil))
	check.Assert(err, gocheck.FitsTypeOf, resources.NoSuchDebugInstanceError{})

	deletedDebugInstance, err := suite.resourcesConnection.DebugInstances.Read.Get(debugInstance.Id)
	check.Assert(err, gocheck.IsNil)
	check.Assert(deletedDebugInstance.IsDeleted, gocheck.Equals, true)
}
