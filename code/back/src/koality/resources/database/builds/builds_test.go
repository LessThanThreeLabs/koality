package builds_test

import (
	"github.com/LessThanThreeLabs/gocheck"
	"koality/resources"
	"koality/resources/database"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

type BuildsTestSuite struct {
	resourcesConnection *resources.Connection
}

var _ = gocheck.Suite(new(BuildsTestSuite))

func (suite *BuildsTestSuite) SetUpTest(check *gocheck.C) {
	err := database.PopulateDatabase()
	check.Assert(err, gocheck.IsNil)

	suite.resourcesConnection, err = database.New()
	check.Assert(err, gocheck.IsNil)
}

func (suite *BuildsTestSuite) TearDownTest(check *gocheck.C) {
	if suite.resourcesConnection != nil {
		suite.resourcesConnection.Close()
	}
}

func (suite *BuildsTestSuite) TestCreateInvalidBuild(check *gocheck.C) {
	repositories, err := suite.resourcesConnection.Repositories.Read.GetAll()
	check.Assert(err, gocheck.IsNil)

	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"

	_, err = suite.resourcesConnection.Builds.Create.Create(13370, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, true)
	_, ok := err.(resources.NoSuchRepositoryError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected NoSuchRepositoryError when providing invalid repository id"))

	_, err = suite.resourcesConnection.Builds.Create.CreateForSnapshot(firstRepository.Id, 13370, headSha, baseSha, headMessage, headUsername, headEmail, emailToNotify, ref)
	_, ok = err.(resources.NoSuchSnapshotError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected NoSuchSnapshotError"))

	_, err = suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, "badheadsha", baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, true)
	check.Assert(err, gocheck.NotNil, gocheck.Commentf("Expected error after providing invalid head sha"))

	_, err = suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, headSha, "badbasesha", headMessage, headUsername, headEmail, nil, emailToNotify, ref, true)
	check.Assert(err, gocheck.NotNil, gocheck.Commentf("Expected error after providing invalid base sha"))

	_, err = suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, "not-an-email", ref, true)
	check.Assert(err, gocheck.NotNil, gocheck.Commentf("Expected error after providing invalid email to notify"))

	_, err = suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, []byte("this is an invalid patch\ndiff file"), emailToNotify, ref, true)
	check.Assert(err, gocheck.NotNil, gocheck.Commentf("Expected error after providing invalid patch contents"))
}

func shallowCheckBuild(check *gocheck.C, build, expectedBuild *resources.Build) {
	check.Assert(build.RepositoryId, gocheck.Equals, expectedBuild.RepositoryId)
	check.Assert(build.Ref, gocheck.Equals, expectedBuild.Ref)
	check.Assert(build.EmailToNotify, gocheck.Equals, expectedBuild.EmailToNotify)
	check.Assert(build.ShouldMerge, gocheck.Equals, expectedBuild.ShouldMerge)
	check.Assert(build.Changeset.HeadSha, gocheck.Equals, expectedBuild.Changeset.HeadSha)
	check.Assert(build.Changeset.BaseSha, gocheck.Equals, expectedBuild.Changeset.BaseSha)
	check.Assert(build.Changeset.HeadMessage, gocheck.Equals, expectedBuild.Changeset.HeadMessage)
	check.Assert(build.Changeset.HeadUsername, gocheck.Equals, expectedBuild.Changeset.HeadUsername)
	check.Assert(build.Changeset.HeadEmail, gocheck.Equals, expectedBuild.Changeset.HeadEmail)
}

func (suite *BuildsTestSuite) TestCreateBuild(check *gocheck.C) {
	createdEventReceived := make(chan bool, 1)
	var createdEventBuild *resources.Build
	buildCreatedHandler := func(build *resources.Build) {
		createdEventBuild = build
		createdEventReceived <- true
	}
	_, err := suite.resourcesConnection.Builds.Subscription.SubscribeToCreatedEvents(buildCreatedHandler)
	check.Assert(err, gocheck.IsNil)

	repositories, err := suite.resourcesConnection.Repositories.Read.GetAll()
	check.Assert(err, gocheck.IsNil)
	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"
	shouldMerge := true

	build, err := suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	check.Assert(err, gocheck.IsNil)

	shallowCheckBuild(check, build, &resources.Build{
		RepositoryId:  firstRepository.Id,
		Ref:           ref,
		EmailToNotify: emailToNotify,
		ShouldMerge:   shouldMerge,
		Changeset: resources.Changeset{
			HeadSha:      headSha,
			BaseSha:      baseSha,
			HeadMessage:  headMessage,
			HeadUsername: headUsername,
			HeadEmail:    headEmail,
		},
	})

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear build creation event")
	}

	check.Assert(createdEventBuild, gocheck.DeepEquals, build)

	buildAgain, err := suite.resourcesConnection.Builds.Read.Get(build.Id)
	check.Assert(err, gocheck.IsNil)

	shallowCheckBuild(check, buildAgain, build)

	_, err = suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	_, ok := err.(resources.ChangesetAlreadyExistsError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected ChangesetAlreadyExistsError when trying to add build with same changeset params twice"))

	build2, err := suite.resourcesConnection.Builds.Create.CreateFromChangeset(firstRepository.Id, build.Changeset.Id, emailToNotify, ref, shouldMerge)
	check.Assert(err, gocheck.IsNil)

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear build creation event")
	}

	check.Assert(createdEventBuild, gocheck.DeepEquals, build2)
}

func (suite *BuildsTestSuite) TestBuildStatus(check *gocheck.C) {
	buildStatusEventReceived := make(chan bool, 1)
	buildStatusEventId := uint64(0)
	buildStatusEventStatus := ""
	buildStatusUpdatedHandler := func(buildId uint64, status string) {
		buildStatusEventId = buildId
		buildStatusEventStatus = status
		buildStatusEventReceived <- true
	}
	_, err := suite.resourcesConnection.Builds.Subscription.SubscribeToStatusUpdatedEvents(buildStatusUpdatedHandler)
	check.Assert(err, gocheck.IsNil)

	repositories, err := suite.resourcesConnection.Repositories.Read.GetAll()
	check.Assert(err, gocheck.IsNil)
	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"
	shouldMerge := true

	build, err := suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	check.Assert(err, gocheck.IsNil)

	check.Assert(build.Status, gocheck.Equals, "declared", gocheck.Commentf("Expected initial build status to be 'declared'"))

	err = suite.resourcesConnection.Builds.Update.SetStatus(build.Id, "passed")
	check.Assert(err, gocheck.IsNil)

	select {
	case <-buildStatusEventReceived:
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear build status updated event")
	}

	check.Assert(build.Id, gocheck.Equals, buildStatusEventId, gocheck.Commentf("Bad build.Id in status updated event"))
	check.Assert(buildStatusEventStatus, gocheck.Equals, "passed", gocheck.Commentf("Bad build status in status updated event"))

	build, err = suite.resourcesConnection.Builds.Read.Get(build.Id)
	check.Assert(err, gocheck.IsNil)
	check.Assert(build.Status, gocheck.Equals, "passed", gocheck.Commentf("Failed to update build status"))

	err = suite.resourcesConnection.Builds.Update.SetStatus(build.Id, "bad-status")
	_, ok := err.(resources.InvalidBuildStatusError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected InvalidBuildStatusError when trying to set status"))
}

func (suite *BuildsTestSuite) TestBuildTimes(check *gocheck.C) {
	buildStartTimeEventReceived := make(chan bool, 1)
	buildStartTimeEventId := uint64(0)
	buildStartTimeEventTime := time.Now()
	buildStartTimeUpdatedHandler := func(buildId uint64, startTime time.Time) {
		buildStartTimeEventId = buildId
		buildStartTimeEventTime = startTime
		buildStartTimeEventReceived <- true
	}
	_, err := suite.resourcesConnection.Builds.Subscription.SubscribeToStartTimeUpdatedEvents(buildStartTimeUpdatedHandler)
	check.Assert(err, gocheck.IsNil)

	buildEndTimeEventReceived := make(chan bool, 1)
	buildEndTimeEventId := uint64(0)
	buildEndTimeEventTime := time.Now()
	buildEndTimeUpdatedHandler := func(buildId uint64, endTime time.Time) {
		buildEndTimeEventId = buildId
		buildEndTimeEventTime = endTime
		buildEndTimeEventReceived <- true
	}
	_, err = suite.resourcesConnection.Builds.Subscription.SubscribeToEndTimeUpdatedEvents(buildEndTimeUpdatedHandler)
	check.Assert(err, gocheck.IsNil)

	repositories, err := suite.resourcesConnection.Repositories.Read.GetAll()
	check.Assert(err, gocheck.IsNil)
	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"
	shouldMerge := true

	build, err := suite.resourcesConnection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	check.Assert(err, gocheck.IsNil)

	err = suite.resourcesConnection.Builds.Update.SetEndTime(build.Id, time.Now())
	check.Assert(err, gocheck.NotNil, gocheck.Commentf("Expected error when setting end time without start time"))

	err = suite.resourcesConnection.Builds.Update.SetStartTime(build.Id, time.Unix(0, 0))
	check.Assert(err, gocheck.NotNil, gocheck.Commentf("Expected error when setting start time before create time"))

	startTime := time.Now()
	err = suite.resourcesConnection.Builds.Update.SetStartTime(build.Id, startTime)
	check.Assert(err, gocheck.IsNil)

	select {
	case <-buildStartTimeEventReceived:
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear build start time event")
	}

	check.Assert(buildStartTimeEventId, gocheck.Equals, build.Id)
	check.Assert(buildStartTimeEventTime, gocheck.Equals, startTime)

	err = suite.resourcesConnection.Builds.Update.SetStartTime(0, time.Now())
	_, ok := err.(resources.NoSuchBuildError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected NoSuchBuildError when trying to set start time for nonexistent build"))

	err = suite.resourcesConnection.Builds.Update.SetEndTime(build.Id, time.Unix(0, 0))
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected error when setting end time before create time"))

	endTime := time.Now()
	err = suite.resourcesConnection.Builds.Update.SetEndTime(build.Id, endTime)
	check.Assert(err, gocheck.IsNil)

	select {
	case <-buildEndTimeEventReceived:
	case <-time.After(10 * time.Second):
		check.Fatal("Failed to hear build end time event")
	}

	check.Assert(buildEndTimeEventId, gocheck.Equals, build.Id)
	check.Assert(buildEndTimeEventTime, gocheck.Equals, endTime)

	err = suite.resourcesConnection.Builds.Update.SetEndTime(0, time.Now())
	_, ok = err.(resources.NoSuchBuildError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected NoSuchBuildError when trying to set end time for nonexistent build"))
}

func (suite *BuildsTestSuite) TestGetTail(check *gocheck.C) {
	repositories, err := suite.resourcesConnection.Repositories.Read.GetAll()
	check.Assert(err, gocheck.IsNil)
	firstRepository := repositories[0]

	builds, err := suite.resourcesConnection.Builds.Read.GetTail(firstRepository.Id, 0, 1)
	check.Assert(err, gocheck.IsNil)
	check.Assert(builds, gocheck.HasLen, 1)

	firstBuild := builds[0]

	builds, err = suite.resourcesConnection.Builds.Read.GetTail(firstRepository.Id, 1, 4)
	check.Assert(err, gocheck.IsNil)
	check.Assert(builds, gocheck.HasLen, 4)

	check.Assert(firstBuild.Id, gocheck.Not(gocheck.Equals), builds[0].Id, gocheck.Commentf("Same build id twice"))

	builds, err = suite.resourcesConnection.Builds.Read.GetTail(firstRepository.Id, 14, 0)
	check.Assert(err, gocheck.NotNil, gocheck.Commentf("Expected error when requesting 0 builds"))
}

func (suite *BuildsTestSuite) TestGetChangesetFromShas(check *gocheck.C) {
	repositories, err := suite.resourcesConnection.Repositories.Read.GetAll()
	check.Assert(err, gocheck.IsNil)
	firstRepository := repositories[0]

	builds, err := suite.resourcesConnection.Builds.Read.GetTail(firstRepository.Id, 0, 1)
	check.Assert(err, gocheck.IsNil)
	firstBuild := builds[0]

	changeset, err := suite.resourcesConnection.Builds.Read.GetChangesetFromShas(firstBuild.Changeset.HeadSha, firstBuild.Changeset.BaseSha, firstBuild.Changeset.PatchContents)
	check.Assert(err, gocheck.IsNil)

	check.Assert(changeset.HeadSha, gocheck.Equals, firstBuild.Changeset.HeadSha)
	check.Assert(changeset.BaseSha, gocheck.Equals, firstBuild.Changeset.BaseSha)

	_, err = suite.resourcesConnection.Builds.Read.GetChangesetFromShas("some-bad-head-sha", firstBuild.Changeset.BaseSha, firstBuild.Changeset.PatchContents)
	_, ok := err.(resources.NoSuchChangesetError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected NoSuchChangesetError when providing invalid head sha"))

	_, err = suite.resourcesConnection.Builds.Read.GetChangesetFromShas(firstBuild.Changeset.HeadSha, "some-bad-base-sha", firstBuild.Changeset.PatchContents)
	_, ok = err.(resources.NoSuchChangesetError)
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected NoSuchChangesetError when providing invalid base sha"))

	_, err = suite.resourcesConnection.Builds.Read.GetChangesetFromShas(firstBuild.Changeset.HeadSha, firstBuild.Changeset.BaseSha, []byte("some-bad-patch-contents"))
	check.Assert(ok, gocheck.Equals, true, gocheck.Commentf("Expected NoSuchChangesetError when providing invalid patch contents"))
}
