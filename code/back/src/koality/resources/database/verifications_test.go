package database

import (
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidVerification(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
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
	mergeTarget := "refs/heads/master"
	emailToNotify := "koalas@koalitycode.com"

	_, err = connection.Verifications.Create.Create(13370, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if _, ok := err.(resources.NoSuchRepositoryError); !ok {
		test.Fatal("Expected NoSuchRepositoryError when providing invalid repository id")
	}

	_, err = connection.Verifications.Create.CreateForSnapshot(firstRepository.Id, 13370, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if _, ok := err.(resources.NoSuchSnapshotError); !ok {
		test.Fatal("Expected NoSuchSnapshotError")
	}

	_, err = connection.Verifications.Create.Create(firstRepository.Id, "badheadsha", baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err == nil {
		test.Fatal("Expected error after providing invalid head sha")
	}

	_, err = connection.Verifications.Create.Create(firstRepository.Id, headSha, "badbasesha", headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err == nil {
		test.Fatal("Expected error after providing invalid base sha")
	}

	_, err = connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, "not-an-email")
	if err == nil {
		test.Fatal("Expected error after providing invalid email to notify")
	}
}

func TestCreateVerification(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	createdEventReceived := make(chan bool, 1)
	var createdEventVerification *resources.Verification
	verificationCreatedHandler := func(verification *resources.Verification) {
		createdEventVerification = verification
		createdEventReceived <- true
	}
	_, err = connection.Verifications.Subscription.SubscribeToCreatedEvents(verificationCreatedHandler)
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
	mergeTarget := "refs/heads/master"
	emailToNotify := "koalas@koalitycode.com"

	verification, err := connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	if verification.RepositoryId != firstRepository.Id {
		test.Fatal("verification.RepositoryId mismatch")
	} else if verification.SnapshotId != 0 {
		test.Fatal("verification.SnapshotId mismatch")
	} else if verification.MergeTarget != mergeTarget {
		test.Fatal("verification.MergeTarget mismatch")
	} else if verification.EmailToNotify != emailToNotify {
		test.Fatal("verification.EmailToNotify mismatch")
	} else if verification.Changeset.HeadSha != headSha {
		test.Fatal("verification.Changeset.HeadSha mismatch")
	} else if verification.Changeset.BaseSha != baseSha {
		test.Fatal("verification.Changeset.BaseSha mismatch")
	} else if verification.Changeset.HeadMessage != headMessage {
		test.Fatal("verification.Changeset.HeadMessage mismatch")
	} else if verification.Changeset.HeadUsername != headUsername {
		test.Fatal("verification.Changeset.HeadUsername mismatch")
	} else if verification.Changeset.HeadEmail != headEmail {
		test.Fatal("verification.Changeset.HeadEmail mismatch")
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear verification creation event")
	}

	if createdEventVerification.Id != verification.Id {
		test.Fatal("Bad verification.Id in verification creation event")
	} else if createdEventVerification.RepositoryId != verification.RepositoryId {
		test.Fatal("Bad verification.RepositoryId in verification creation event")
	} else if createdEventVerification.SnapshotId != verification.SnapshotId {
		test.Fatal("Bad verification.SnapshotId in verification creation event")
	} else if createdEventVerification.MergeTarget != verification.MergeTarget {
		test.Fatal("Bad verification.MergeTarget in verification creation event")
	} else if createdEventVerification.EmailToNotify != verification.EmailToNotify {
		test.Fatal("Bad verification.EmailToNotify in verification creation event")
	} else if createdEventVerification.Changeset.HeadSha != verification.Changeset.HeadSha {
		test.Fatal("Bad verification.Changeset.HeadSha in verification creation event")
	} else if createdEventVerification.Changeset.BaseSha != verification.Changeset.BaseSha {
		test.Fatal("Bad verification.Changeset.BaseSha in verification creation event")
	} else if createdEventVerification.Changeset.HeadMessage != verification.Changeset.HeadMessage {
		test.Fatal("Bad verification.Changeset.HeadMessage in verification creation event")
	} else if createdEventVerification.Changeset.HeadUsername != verification.Changeset.HeadUsername {
		test.Fatal("Bad verification.Changeset.HeadUsername in verification creation event")
	} else if createdEventVerification.Changeset.HeadEmail != verification.Changeset.HeadEmail {
		test.Fatal("Bad verification.Changeset.HeadEmail in verification creation event")
	}

	verificationAgain, err := connection.Verifications.Read.Get(verification.Id)
	if err != nil {
		test.Fatal(err)
	}

	if verification.RepositoryId != verificationAgain.RepositoryId {
		test.Fatal("verification.RepositoryId mismatch")
	} else if verification.SnapshotId != verificationAgain.SnapshotId {
		test.Fatal("verification.SnapshotId mismatch")
	} else if verification.MergeTarget != verificationAgain.MergeTarget {
		test.Fatal("verification.MergeTarget mismatch")
	} else if verification.EmailToNotify != verificationAgain.EmailToNotify {
		test.Fatal("verification.EmailToNotify mismatch")
	} else if verification.Changeset.HeadSha != verificationAgain.Changeset.HeadSha {
		test.Fatal("verification.Changeset.HeadSha mismatch")
	} else if verification.Changeset.BaseSha != verificationAgain.Changeset.BaseSha {
		test.Fatal("verification.Changeset.BaseSha mismatch")
	} else if verification.Changeset.HeadMessage != verificationAgain.Changeset.HeadMessage {
		test.Fatal("verification.Changeset.HeadMessage mismatch")
	} else if verification.Changeset.HeadUsername != verificationAgain.Changeset.HeadUsername {
		test.Fatal("verification.Changeset.HeadUsername mismatch")
	} else if verification.Changeset.HeadEmail != verificationAgain.Changeset.HeadEmail {
		test.Fatal("verification.Changeset.HeadEmail mismatch")
	}

	_, err = connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if _, ok := err.(resources.ChangesetAlreadyExistsError); !ok {
		test.Fatal("Expected ChangesetAlreadyExistsError when trying to add verification with same changeset params twice")
	}

	verification2, err := connection.Verifications.Create.CreateFromChangeset(firstRepository.Id, verification.Changeset.Id, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear verification creation event")
	}

	if createdEventVerification.Id != verification2.Id {
		test.Fatal("Bad verification.Id in verification creation event")
	} else if createdEventVerification.RepositoryId != verification2.RepositoryId {
		test.Fatal("Bad verification.RepositoryId in verification creation event")
	} else if createdEventVerification.SnapshotId != verification2.SnapshotId {
		test.Fatal("Bad verification.SnapshotId in verification creation event")
	} else if createdEventVerification.MergeTarget != verification2.MergeTarget {
		test.Fatal("Bad verification.MergeTarget in verification creation event")
	} else if createdEventVerification.EmailToNotify != verification2.EmailToNotify {
		test.Fatal("Bad verification.EmailToNotify in verification creation event")
	} else if createdEventVerification.Changeset.HeadSha != verification2.Changeset.HeadSha {
		test.Fatal("Bad verification.Changeset.HeadSha in verification creation event")
	} else if createdEventVerification.Changeset.BaseSha != verification2.Changeset.BaseSha {
		test.Fatal("Bad verification.Changeset.BaseSha in verification creation event")
	} else if createdEventVerification.Changeset.HeadMessage != verification2.Changeset.HeadMessage {
		test.Fatal("Bad verification.Changeset.HeadMessage in verification creation event")
	} else if createdEventVerification.Changeset.HeadUsername != verification2.Changeset.HeadUsername {
		test.Fatal("Bad verification.Changeset.HeadUsername in verification creation event")
	} else if createdEventVerification.Changeset.HeadEmail != verification2.Changeset.HeadEmail {
		test.Fatal("Bad verification.Changeset.HeadEmail in verification creation event")
	}
}

func TestVerificationStatuses(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	verificationStatusEventReceived := make(chan bool, 1)
	verificationStatusEventId := uint64(0)
	verificationStatusEventStatus := ""
	verificationStatusUpdatedHandler := func(verificationId uint64, status string) {
		verificationStatusEventId = verificationId
		verificationStatusEventStatus = status
		verificationStatusEventReceived <- true
	}
	_, err = connection.Verifications.Subscription.SubscribeToStatusUpdatedEvents(verificationStatusUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	verificationMergeStatusEventReceived := make(chan bool, 1)
	verificationMergeStatusEventId := uint64(0)
	verificationMergeStatusEventStatus := ""
	verificationMergeStatusUpdatedHandler := func(verificationId uint64, mergeStatus string) {
		verificationMergeStatusEventId = verificationId
		verificationMergeStatusEventStatus = mergeStatus
		verificationMergeStatusEventReceived <- true
	}
	_, err = connection.Verifications.Subscription.SubscribeToMergeStatusUpdatedEvents(verificationMergeStatusUpdatedHandler)
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
	mergeTarget := "refs/heads/master"
	emailToNotify := "koalas@koalitycode.com"

	verification, err := connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	if verification.Status != "declared" {
		test.Fatal("Expected initial verification status to be 'declared'")
	}

	err = connection.Verifications.Update.SetStatus(verification.Id, "passed")
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-verificationStatusEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear verification status updated event")
	}

	if verificationStatusEventId != verification.Id {
		test.Fatal("Bad verification.Id in status updated event")
	} else if verificationStatusEventStatus != "passed" {
		test.Fatal("Bad verification status in status updated event")
	}

	verification, err = connection.Verifications.Read.Get(verification.Id)
	if err != nil {
		test.Fatal(err)
	} else if verification.Status != "passed" {
		test.Fatal("Failed to update verification status")
	}

	err = connection.Verifications.Update.SetStatus(verification.Id, "bad-status")
	if _, ok := err.(resources.InvalidVerificationStatusError); !ok {
		test.Fatal("Expected InvalidVerificationStatusError when trying to set status")
	}

	err = connection.Verifications.Update.SetMergeStatus(verification.Id, "failed")
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-verificationMergeStatusEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear verification merge status updated event")
	}

	if verificationMergeStatusEventId != verification.Id {
		test.Fatal("Bad verification.Id in merge status updated event")
	} else if verificationMergeStatusEventStatus != "failed" {
		test.Fatal("Bad verification merge status in merge status updated event")
	}

	verification, err = connection.Verifications.Read.Get(verification.Id)
	if err != nil {
		test.Fatal(err)
	} else if verification.MergeStatus != "failed" {
		test.Fatal("Failed to update verification merge status")
	}

	err = connection.Verifications.Update.SetMergeStatus(verification.Id, "bad-merge-status")
	if _, ok := err.(resources.InvalidVerificationMergeStatusError); !ok {
		test.Fatal("Expected InvalidVerificationMergeStatusError when trying to set merge status")
	}
}

func TestVerificationTimes(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	verificationStartTimeEventReceived := make(chan bool, 1)
	verificationStartTimeEventId := uint64(0)
	verificationStartTimeEventTime := time.Now()
	verificationStartTimeUpdatedHandler := func(verificationId uint64, startTime time.Time) {
		verificationStartTimeEventId = verificationId
		verificationStartTimeEventTime = startTime
		verificationStartTimeEventReceived <- true
	}
	_, err = connection.Verifications.Subscription.SubscribeToStartTimeUpdatedEvents(verificationStartTimeUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	verificationEndTimeEventReceived := make(chan bool, 1)
	verificationEndTimeEventId := uint64(0)
	verificationEndTimeEventTime := time.Now()
	verificationEndTimeUpdatedHandler := func(verificationId uint64, endTime time.Time) {
		verificationEndTimeEventId = verificationId
		verificationEndTimeEventTime = endTime
		verificationEndTimeEventReceived <- true
	}
	_, err = connection.Verifications.Subscription.SubscribeToEndTimeUpdatedEvents(verificationEndTimeUpdatedHandler)
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
	mergeTarget := "refs/heads/master"
	emailToNotify := "koalas@koalitycode.com"

	verification, err := connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Verifications.Update.SetStartTime(verification.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	startTime := time.Now()
	err = connection.Verifications.Update.SetStartTime(verification.Id, startTime)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-verificationStartTimeEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear verification start time event")
	}

	if verificationStartTimeEventId != verification.Id {
		test.Fatal("Bad verification.Id in start time event")
	} else if verificationStartTimeEventTime != startTime {
		test.Fatal("Bad verification start time in start time event")
	}

	err = connection.Verifications.Update.SetStartTime(0, time.Now())
	if _, ok := err.(resources.NoSuchVerificationError); !ok {
		test.Fatal("Expected NoSuchVerificationError when trying to set start time for nonexistent verification")
	}

	err = connection.Verifications.Update.SetEndTime(verification.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	endTime := time.Now()
	err = connection.Verifications.Update.SetEndTime(verification.Id, endTime)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-verificationEndTimeEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear verification end time event")
	}

	if verificationEndTimeEventId != verification.Id {
		test.Fatal("Bad verification.Id in end time event")
	} else if verificationEndTimeEventTime != endTime {
		test.Fatal("Bad verification end time in end time event")
	}

	err = connection.Verifications.Update.SetEndTime(0, time.Now())
	if _, ok := err.(resources.NoSuchVerificationError); !ok {
		test.Fatal("Expected NoSuchVerificationError when trying to set end time for nonexistent verification")
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

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}
	firstRepository := repositories[0]

	verifications, err := connection.Verifications.Read.GetTail(firstRepository.Id, 0, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(verifications) != 1 {
		test.Fatal("Expected only one verification")
	}

	firstVerification := verifications[0]

	verifications, err = connection.Verifications.Read.GetTail(firstRepository.Id, 1, 4)
	if err != nil {
		test.Fatal(err)
	} else if len(verifications) != 4 {
		test.Fatal("Expected four verifications")
	}

	if firstVerification.Id == verifications[0].Id {
		test.Fatal("Same verification id twice")
	}

	verifications, err = connection.Verifications.Read.GetTail(firstRepository.Id, 14, 0)
	if err == nil {
		test.Fatal("Expected error when requesting 0 verifications")
	}
}
