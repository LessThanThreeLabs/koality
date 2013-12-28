package database

import (
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidVerification(test *testing.T) {
	PopulateDatabase()

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
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	verificationCreatedEventReceived := make(chan bool, 1)
	verificationCreatedEventId := uint64(0)
	verificationCreatedHandler := func(verificationId uint64) {
		verificationCreatedEventId = verificationId
		verificationCreatedEventReceived <- true
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

	verificationId, err := connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-verificationCreatedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear verification creation event")
	}

	if verificationCreatedEventId != verificationId {
		test.Fatal("Bad verificationId in verification creation event")
	}

	verification, err := connection.Verifications.Read.Get(verificationId)
	if err != nil {
		test.Fatal(err)
	} else if verification.Id != verificationId {
		test.Fatal("verification.Id mismatch")
	}

	_, err = connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if _, ok := err.(resources.ChangesetAlreadyExistsError); !ok {
		test.Fatal("Expected ChangesetAlreadyExistsError when trying to add verification with same changeset params twice")
	}

	verificationId2, err := connection.Verifications.Create.CreateFromChangeset(firstRepository.Id, verification.Changeset.Id, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-verificationCreatedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear verification creation event")
	}

	if verificationCreatedEventId != verificationId2 {
		test.Fatal("Bad verificationId in verification creation event")
	}

	verification2, err := connection.Verifications.Read.Get(verificationId2)
	if err != nil {
		test.Fatal(err)
	} else if verification2.Id != verificationId2 {
		test.Fatal("verification.Id mismatch")
	}
}

func TestVerificationStatuses(test *testing.T) {
	PopulateDatabase()

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

	verificationId, err := connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	verification, err := connection.Verifications.Read.Get(verificationId)
	if err != nil {
		test.Fatal(err)
	} else if verification.Status != "declared" {
		test.Fatal("Expected initial verification status to be 'declared'")
	}

	err = connection.Verifications.Update.SetStatus(verificationId, "passed")
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-verificationStatusEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear verification status updated event")
	}

	if verificationStatusEventId != verificationId {
		test.Fatal("Bad verificationId in status updated event")
	} else if verificationStatusEventStatus != "passed" {
		test.Fatal("Bad verification status in status updated event")
	}

	verification, err = connection.Verifications.Read.Get(verificationId)
	if err != nil {
		test.Fatal(err)
	} else if verification.Status != "passed" {
		test.Fatal("Failed to update verification status")
	}

	err = connection.Verifications.Update.SetStatus(verificationId, "bad-status")
	if _, ok := err.(resources.InvalidVerificationStatusError); !ok {
		test.Fatal("Expected InvalidVerificationStatusError when trying to set status")
	}

	err = connection.Verifications.Update.SetMergeStatus(verificationId, "failed")
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-verificationMergeStatusEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear verification merge status updated event")
	}

	if verificationMergeStatusEventId != verificationId {
		test.Fatal("Bad verificationId in merge status updated event")
	} else if verificationMergeStatusEventStatus != "failed" {
		test.Fatal("Bad verification merge status in merge status updated event")
	}

	verification, err = connection.Verifications.Read.Get(verificationId)
	if err != nil {
		test.Fatal(err)
	} else if verification.MergeStatus != "failed" {
		test.Fatal("Failed to update verification merge status")
	}

	err = connection.Verifications.Update.SetMergeStatus(verificationId, "bad-merge-status")
	if _, ok := err.(resources.InvalidVerificationMergeStatusError); !ok {
		test.Fatal("Expected InvalidVerificationMergeStatusError when trying to set merge status")
	}
}

func TestVerificationTimes(test *testing.T) {
	PopulateDatabase()

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

	verificationId, err := connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Verifications.Update.SetEndTime(verificationId, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Verifications.Update.SetStartTime(verificationId, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	startTime := time.Now()
	err = connection.Verifications.Update.SetStartTime(verificationId, startTime)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-verificationStartTimeEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear verification start time event")
	}

	if verificationStartTimeEventId != verificationId {
		test.Fatal("Bad verificationId in start time event")
	} else if verificationStartTimeEventTime != startTime {
		test.Fatal("Bad verification start time in start time event")
	}

	err = connection.Verifications.Update.SetStartTime(0, time.Now())
	if _, ok := err.(resources.NoSuchVerificationError); !ok {
		test.Fatal("Expected NoSuchVerificationError when trying to set start time for nonexistent verification")
	}

	err = connection.Verifications.Update.SetEndTime(verificationId, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	endTime := time.Now()
	err = connection.Verifications.Update.SetEndTime(verificationId, endTime)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-verificationEndTimeEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear verification end time event")
	}

	if verificationEndTimeEventId != verificationId {
		test.Fatal("Bad verificationId in end time event")
	} else if verificationEndTimeEventTime != endTime {
		test.Fatal("Bad verification end time in end time event")
	}

	err = connection.Verifications.Update.SetEndTime(0, time.Now())
	if _, ok := err.(resources.NoSuchVerificationError); !ok {
		test.Fatal("Expected NoSuchVerificationError when trying to set end time for nonexistent verification")
	}
}

func TestGetTail(test *testing.T) {
	PopulateDatabase()

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
