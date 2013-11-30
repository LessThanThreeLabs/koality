package database

import (
	"koality/resources"
	"testing"
	"time"
)

const (
	verificationRepositoryName      = "repository-name"
	verificationRepositoryVcsType   = "git"
	verificationRepositoryLocalUri  = "git@local.uri.com:name.git"
	verificationRepositoryRemoteUri = "git@remote.uri.com:name.git"
	headSha                         = "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha                         = "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage                     = "This is an awesome commit message"
	headUsername                    = "Jordan Potter"
	headEmail                       = "jpotter@koalitycode.com"
	mergeTarget                     = "refs/heads/master"
	emailToNotify                   = "koalas@koalitycode.com"
)

var (
	connection   *resources.Connection
	repositoryId uint64
)

// TODO: get rid of this, import database dump instead
func TestPrepareOtherTests(test *testing.T) {
	var err error
	connection, err = New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err = connection.Repositories.Create.Create(verificationRepositoryName, verificationRepositoryVcsType, verificationRepositoryLocalUri, verificationRepositoryRemoteUri)
	if err != nil {
		test.Fatal(err)
	}
}

func TestCreateInvalidVerification(test *testing.T) {
	_, err := connection.Verifications.Create.Create(0, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err == nil {
		test.Fatal("Expected error after providing invalid repository id")
	}

	_, err = connection.Verifications.Create.Create(repositoryId, "badheadsha", baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err == nil {
		test.Fatal("Expected error after providing invalid head sha")
	}

	_, err = connection.Verifications.Create.Create(repositoryId, headSha, "badbasesha", headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err == nil {
		test.Fatal("Expected error after providing invalid base sha")
	}

	_, err = connection.Verifications.Create.Create(repositoryId, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, "not-an-email")
	if err == nil {
		test.Fatal("Expected error after providing invalid email to notify")
	}
}

func TestCreateVerification(test *testing.T) {
	verificationId, err := connection.Verifications.Create.Create(repositoryId, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
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

	_, err = connection.Verifications.Create.Create(repositoryId, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
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
