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
	}

	if verification.Id != verificationId {
		test.Fatal("verification.Id mismatch")
	}

	_, err = connection.Verifications.Create.Create(firstRepository.Id, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if _, ok := err.(resources.ChangesetAlreadyExistsError); !ok {
		test.Fatal("Expected ChangesetAlreadyExistsError when trying to add verification with same changeset params twice")
	}

	err = connection.Verifications.Update.SetStatus(verificationId, "passed")
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Verifications.Update.SetStatus(verificationId, "bad-status")
	if _, ok := err.(resources.InvalidVerificationStatusError); !ok {
		test.Fatal("Expected InvalidVerificationStatusError when trying to set status")
	}

	err = connection.Verifications.Update.SetMergeStatus(verificationId, "passed")
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Verifications.Update.SetMergeStatus(verificationId, "bad-merge-status")
	if _, ok := err.(resources.InvalidVerificationMergeStatusError); !ok {
		test.Fatal("Expected InvalidVerificationMergeStatusError when trying to set merge status")
	}

	err = connection.Verifications.Update.SetEndTime(verificationId, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Verifications.Update.SetStartTime(verificationId, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	err = connection.Verifications.Update.SetStartTime(verificationId, time.Now())
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Verifications.Update.SetEndTime(verificationId, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	err = connection.Verifications.Update.SetEndTime(verificationId, time.Now())
	if err != nil {
		test.Fatal(err)
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
