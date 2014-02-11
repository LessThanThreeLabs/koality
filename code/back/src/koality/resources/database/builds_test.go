package database

import (
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidBuild(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"

	_, err = connection.Builds.Create.Create(13370, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, true)
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when providing invalid repository id")
	}

	_, err = connection.Builds.Create.CreateForSnapshot(firstRepository.Id, 13370, headSha, baseSha, headMessage, headUsername, headEmail, emailToNotify, ref)
	if _, ok := err.(resources.NoSuchSnapshotError); !ok {
		test.Fatal("Expected NoSuchSnapshotError")
	}

	_, err = connection.Builds.Create.Create(firstRepository.Id, "badheadsha", baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, true)
	if err == nil {
		test.Fatal("Expected error after providing invalid head sha")
	}

	_, err = connection.Builds.Create.Create(firstRepository.Id, headSha, "badbasesha", headMessage, headUsername, headEmail, nil, emailToNotify, ref, true)
	if err == nil {
		test.Fatal("Expected error after providing invalid base sha")
	}

	_, err = connection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, "not-an-email", ref, true)
	if err == nil {
		test.Fatal("Expected error after providing invalid email to notify")
	}

	_, err = connection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, []byte("this is an invalid patch\ndiff file"), emailToNotify, ref, true)
	if err == nil {
		test.Fatal("Expected error after providing invalid patch contents")
	}
}

func TestCreateBuild(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	createdEventReceived := make(chan bool, 1)
	var createdEventBuild *resources.Build
	buildCreatedHandler := func(build *resources.Build) {
		createdEventBuild = build
		createdEventReceived <- true
	}
	_, err = connection.Builds.Subscription.SubscribeToCreatedEvents(buildCreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}
	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"
	shouldMerge := true

	build, err := connection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	if err != nil {
		test.Fatal(err)
	}

	if build.RepositoryId != firstRepository.Id {
		test.Fatal("build.RepositoryId mismatch")
	} else if build.Ref != ref {
		test.Fatal("build.Ref mismatch")
	} else if build.EmailToNotify != emailToNotify {
		test.Fatal("build.EmailToNotify mismatch")
	} else if build.ShouldMerge != shouldMerge {
		test.Fatal("build.ShouldMerge mismatch")
	} else if build.Changeset.HeadSha != headSha {
		test.Fatal("build.Changeset.HeadSha mismatch")
	} else if build.Changeset.BaseSha != baseSha {
		test.Fatal("build.Changeset.BaseSha mismatch")
	} else if build.Changeset.HeadMessage != headMessage {
		test.Fatal("build.Changeset.HeadMessage mismatch")
	} else if build.Changeset.HeadUsername != headUsername {
		test.Fatal("build.Changeset.HeadUsername mismatch")
	} else if build.Changeset.HeadEmail != headEmail {
		test.Fatal("build.Changeset.HeadEmail mismatch")
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear build creation event")
	}

	if createdEventBuild.Id != build.Id {
		test.Fatal("Bad build.Id in build creation event")
	} else if createdEventBuild.RepositoryId != build.RepositoryId {
		test.Fatal("Bad build.RepositoryId in build creation event")
	} else if createdEventBuild.Ref != build.Ref {
		test.Fatal("Bad build.Ref in build creation event")
	} else if createdEventBuild.ShouldMerge != build.ShouldMerge {
		test.Fatal("Bad build.ShouldMerge in build creation event")
	} else if createdEventBuild.EmailToNotify != build.EmailToNotify {
		test.Fatal("Bad build.EmailToNotify in build creation event")
	} else if createdEventBuild.Changeset.HeadSha != build.Changeset.HeadSha {
		test.Fatal("Bad build.Changeset.HeadSha in build creation event")
	} else if createdEventBuild.Changeset.BaseSha != build.Changeset.BaseSha {
		test.Fatal("Bad build.Changeset.BaseSha in build creation event")
	} else if createdEventBuild.Changeset.HeadMessage != build.Changeset.HeadMessage {
		test.Fatal("Bad build.Changeset.HeadMessage in build creation event")
	} else if createdEventBuild.Changeset.HeadUsername != build.Changeset.HeadUsername {
		test.Fatal("Bad build.Changeset.HeadUsername in build creation event")
	} else if createdEventBuild.Changeset.HeadEmail != build.Changeset.HeadEmail {
		test.Fatal("Bad build.Changeset.HeadEmail in build creation event")
	}

	buildAgain, err := connection.Builds.Read.Get(build.Id)
	if err != nil {
		test.Fatal(err)
	}

	if build.RepositoryId != buildAgain.RepositoryId {
		test.Fatal("build.RepositoryId mismatch")
	} else if build.Ref != buildAgain.Ref {
		test.Fatal("build.Ref mismatch")
	} else if build.ShouldMerge != buildAgain.ShouldMerge {
		test.Fatal("build.ShouldMerge mismatch")
	} else if build.EmailToNotify != buildAgain.EmailToNotify {
		test.Fatal("build.EmailToNotify mismatch")
	} else if build.Changeset.HeadSha != buildAgain.Changeset.HeadSha {
		test.Fatal("build.Changeset.HeadSha mismatch")
	} else if build.Changeset.BaseSha != buildAgain.Changeset.BaseSha {
		test.Fatal("build.Changeset.BaseSha mismatch")
	} else if build.Changeset.HeadMessage != buildAgain.Changeset.HeadMessage {
		test.Fatal("build.Changeset.HeadMessage mismatch")
	} else if build.Changeset.HeadUsername != buildAgain.Changeset.HeadUsername {
		test.Fatal("build.Changeset.HeadUsername mismatch")
	} else if build.Changeset.HeadEmail != buildAgain.Changeset.HeadEmail {
		test.Fatal("build.Changeset.HeadEmail mismatch")
	}

	_, err = connection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	if _, ok := err.(resources.ChangesetAlreadyExistsError); !ok {
		test.Fatal("Expected ChangesetAlreadyExistsError when trying to add build with same changeset params twice")
	}

	build2, err := connection.Builds.Create.CreateFromChangeset(firstRepository.Id, build.Changeset.Id, emailToNotify, ref, shouldMerge)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear build creation event")
	}

	if createdEventBuild.Id != build2.Id {
		test.Fatal("Bad build.Id in build creation event")
	} else if createdEventBuild.RepositoryId != build2.RepositoryId {
		test.Fatal("Bad build.RepositoryId in build creation event")
	} else if createdEventBuild.Ref != build2.Ref {
		test.Fatal("Bad build.Ref in build creation event")
	} else if createdEventBuild.ShouldMerge != build2.ShouldMerge {
		test.Fatal("Bad build.ShouldMerge in build creation event")
	} else if createdEventBuild.EmailToNotify != build2.EmailToNotify {
		test.Fatal("Bad build.EmailToNotify in build creation event")
	} else if createdEventBuild.Changeset.HeadSha != build2.Changeset.HeadSha {
		test.Fatal("Bad build.Changeset.HeadSha in build creation event")
	} else if createdEventBuild.Changeset.BaseSha != build2.Changeset.BaseSha {
		test.Fatal("Bad build.Changeset.BaseSha in build creation event")
	} else if createdEventBuild.Changeset.HeadMessage != build2.Changeset.HeadMessage {
		test.Fatal("Bad build.Changeset.HeadMessage in build creation event")
	} else if createdEventBuild.Changeset.HeadUsername != build2.Changeset.HeadUsername {
		test.Fatal("Bad build.Changeset.HeadUsername in build creation event")
	} else if createdEventBuild.Changeset.HeadEmail != build2.Changeset.HeadEmail {
		test.Fatal("Bad build.Changeset.HeadEmail in build creation event")
	}
}

func TestBuildStatus(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	buildStatusEventReceived := make(chan bool, 1)
	buildStatusEventId := uint64(0)
	buildStatusEventStatus := ""
	buildStatusUpdatedHandler := func(buildId uint64, status string) {
		buildStatusEventId = buildId
		buildStatusEventStatus = status
		buildStatusEventReceived <- true
	}
	_, err = connection.Builds.Subscription.SubscribeToStatusUpdatedEvents(buildStatusUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}
	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"
	shouldMerge := true

	build, err := connection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	if err != nil {
		test.Fatal(err)
	}

	if build.Status != "declared" {
		test.Fatal("Expected initial build status to be 'declared'")
	}

	err = connection.Builds.Update.SetStatus(build.Id, "passed")
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-buildStatusEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear build status updated event")
	}

	if buildStatusEventId != build.Id {
		test.Fatal("Bad build.Id in status updated event")
	} else if buildStatusEventStatus != "passed" {
		test.Fatal("Bad build status in status updated event")
	}

	build, err = connection.Builds.Read.Get(build.Id)
	if err != nil {
		test.Fatal(err)
	} else if build.Status != "passed" {
		test.Fatal("Failed to update build status")
	}

	err = connection.Builds.Update.SetStatus(build.Id, "bad-status")
	if _, ok := err.(resources.InvalidBuildStatusError); !ok {
		test.Fatal("Expected InvalidBuildStatusError when trying to set status")
	}
}

func TestBuildTimes(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	buildStartTimeEventReceived := make(chan bool, 1)
	buildStartTimeEventId := uint64(0)
	buildStartTimeEventTime := time.Now()
	buildStartTimeUpdatedHandler := func(buildId uint64, startTime time.Time) {
		buildStartTimeEventId = buildId
		buildStartTimeEventTime = startTime
		buildStartTimeEventReceived <- true
	}
	_, err = connection.Builds.Subscription.SubscribeToStartTimeUpdatedEvents(buildStartTimeUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	buildEndTimeEventReceived := make(chan bool, 1)
	buildEndTimeEventId := uint64(0)
	buildEndTimeEventTime := time.Now()
	buildEndTimeUpdatedHandler := func(buildId uint64, endTime time.Time) {
		buildEndTimeEventId = buildId
		buildEndTimeEventTime = endTime
		buildEndTimeEventReceived <- true
	}
	_, err = connection.Builds.Subscription.SubscribeToEndTimeUpdatedEvents(buildEndTimeUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}
	firstRepository := repositories[0]

	headSha := "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha := "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage := "This is an awesome commit message"
	headUsername := "Jordan Potter"
	headEmail := "jpotter@koalitycode.com"
	emailToNotify := "koalas@koalitycode.com"
	ref := "refs/heads/master"
	shouldMerge := true

	build, err := connection.Builds.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, emailToNotify, ref, shouldMerge)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Builds.Update.SetEndTime(build.Id, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Builds.Update.SetStartTime(build.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	startTime := time.Now()
	err = connection.Builds.Update.SetStartTime(build.Id, startTime)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-buildStartTimeEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear build start time event")
	}

	if buildStartTimeEventId != build.Id {
		test.Fatal("Bad build.Id in start time event")
	} else if buildStartTimeEventTime != startTime {
		test.Fatal("Bad build start time in start time event")
	}

	err = connection.Builds.Update.SetStartTime(0, time.Now())
	if _, ok := err.(resources.NoSuchBuildError); !ok {
		test.Fatal("Expected NoSuchBuildError when trying to set start time for nonexistent build")
	}

	err = connection.Builds.Update.SetEndTime(build.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	endTime := time.Now()
	err = connection.Builds.Update.SetEndTime(build.Id, endTime)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-buildEndTimeEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear build end time event")
	}

	if buildEndTimeEventId != build.Id {
		test.Fatal("Bad build.Id in end time event")
	} else if buildEndTimeEventTime != endTime {
		test.Fatal("Bad build end time in end time event")
	}

	err = connection.Builds.Update.SetEndTime(0, time.Now())
	if _, ok := err.(resources.NoSuchBuildError); !ok {
		test.Fatal("Expected NoSuchBuildError when trying to set end time for nonexistent build")
	}
}

func TestGetTail(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}
	firstRepository := repositories[0]

	builds, err := connection.Builds.Read.GetTail(firstRepository.Id, 0, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(builds) != 1 {
		test.Fatal("Expected only one build")
	}

	firstBuild := builds[0]

	builds, err = connection.Builds.Read.GetTail(firstRepository.Id, 1, 4)
	if err != nil {
		test.Fatal(err)
	} else if len(builds) != 4 {
		test.Fatal("Expected four builds")
	}

	if firstBuild.Id == builds[0].Id {
		test.Fatal("Same build id twice")
	}

	builds, err = connection.Builds.Read.GetTail(firstRepository.Id, 14, 0)
	if err == nil {
		test.Fatal("Expected error when requesting 0 builds")
	}
}

func TestGetChangesetFromShas(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}
	firstRepository := repositories[0]

	builds, err := connection.Builds.Read.GetTail(firstRepository.Id, 0, 1)
	if err != nil {
		test.Fatal(err)
	}
	firstBuild := builds[0]

	changeset, err := connection.Builds.Read.GetChangesetFromShas(firstBuild.Changeset.HeadSha, firstBuild.Changeset.BaseSha, firstBuild.Changeset.PatchContents)
	if err != nil {
		test.Fatal(err)
	}

	if changeset.HeadSha != firstBuild.Changeset.HeadSha {
		test.Fatal("changeset.HeadSha mismatch")
	} else if changeset.BaseSha != firstBuild.Changeset.BaseSha {
		test.Fatal("changeset.BaseSha mismatch")
	}

	_, err = connection.Builds.Read.GetChangesetFromShas("some-bad-head-sha", firstBuild.Changeset.BaseSha, firstBuild.Changeset.PatchContents)
	if _, ok := err.(resources.NoSuchChangesetError); !ok {
		test.Fatal("Expected NoSuchChangesetError when providing invalid head sha")
	}

	_, err = connection.Builds.Read.GetChangesetFromShas(firstBuild.Changeset.HeadSha, "some-bad-base-sha", firstBuild.Changeset.PatchContents)
	if _, ok := err.(resources.NoSuchChangesetError); !ok {
		test.Fatal("Expected NoSuchChangesetError when providing invalid base sha")
	}

	_, err = connection.Builds.Read.GetChangesetFromShas(firstBuild.Changeset.HeadSha, firstBuild.Changeset.BaseSha, []byte("some-bad-patch-contents"))
	if _, ok := err.(resources.NoSuchChangesetError); !ok {
		test.Fatal("Expected NoSuchChangesetError when providing invalid patch contents")
	}
}
