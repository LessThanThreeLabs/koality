package database

import (
	"fmt"
	"koality/resources"
	"testing"
	"time"
)

const (
	stageRepositoryName       = "stage-repository-name"
	stageRepositoryVcsType    = "git"
	stageRepositoryLocalUri   = "git@local.uri.com:stage-name.git"
	stageRepositoryRemoteUri  = "git@remote.uri.com:stage-name.git"
	verificationHeadSha       = "a5a1134e5ca1050a2ea01b1b8aaf945bc758ec49"
	verificationBaseSha       = "5984b365f6a7287d8b3674b200525bb769a5a3de"
	verificationHeadMessage   = "This is an awesome commit message"
	verificationHeadUsername  = "Jordan Potter"
	verificationHeadEmail     = "jpotter@koalitycode.com"
	verificationMergeTarget   = "refs/heads/master"
	verificationEmailToNotify = "koalas@koalitycode.com"
	stageName                 = "awesome stage"
	stageFlavor               = "test"
	stageOrderNumber          = 17
)

var (
	// connection          *resources.Connection
	stageRepositoryId   uint64
	stageVerificationId uint64
)

// TODO: get rid of this, import database dump instead
func TestStagesPrepareOtherTests(test *testing.T) {
	var err error
	connection, err = New()
	if err != nil {
		test.Fatal(err)
	}

	stageRepositoryId, err = connection.Repositories.Create.Create(stageRepositoryName, stageRepositoryVcsType, stageRepositoryLocalUri, stageRepositoryRemoteUri)
	if err != nil {
		test.Fatal(err)
	}

	stageVerificationId, err = connection.Verifications.Create.Create(stageRepositoryId, verificationHeadSha, verificationBaseSha, verificationHeadMessage, verificationHeadUsername, verificationHeadEmail, verificationMergeTarget, verificationEmailToNotify)
	if err != nil {
		test.Fatal(err)
	}
}

func TestCreateInvalidStage(test *testing.T) {
	_, err := connection.Stages.Create.Create(0, stageName, stageFlavor, stageOrderNumber)
	if _, ok := err.(resources.NoSuchVerificationError); !ok {
		test.Fatal("Expected NoSuchVerificationError when providing invalid verification id")
	}
	_, err = connection.Stages.Create.Create(stageVerificationId, "", stageFlavor, stageOrderNumber)
	if err == nil {
		test.Fatal("Expected error after providing invalid stage name")
	}

	_, err = connection.Stages.Create.Create(stageVerificationId, stageName, "bad-flavor", stageOrderNumber)
	if err == nil {
		test.Fatal("Expected error after providing invalid flavor")
	}
}

func TestCreateStage(test *testing.T) {
	stageId, err := connection.Stages.Create.Create(stageVerificationId, stageName, stageFlavor, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stage, err := connection.Stages.Read.Get(stageId)
	if err != nil {
		test.Fatal(err)
	} else if stage.Id != stageId {
		test.Fatal("stage.Id mismatch")
	}

	_, err = connection.Stages.Create.Create(stageVerificationId, stageName, stageFlavor, stageOrderNumber)
	if _, ok := err.(resources.StageAlreadyExistsError); !ok {
		test.Fatal("Expected StageAlreadyExistsError when trying to add same stage twice")
	}

	stageRun1Id, err := connection.Stages.Create.CreateRun(stageId)
	if err != nil {
		test.Fatal(err)
	}

	stageRun2Id, err := connection.Stages.Create.CreateRun(stageId)
	if err != nil {
		test.Fatal(err)
	}

	stage, err = connection.Stages.Read.Get(stageId)
	if err != nil {
		test.Fatal(err)
	}

	stageRun1Found, stageRun2Found := false, false
	for _, stageRun := range stage.Runs {
		if stageRun.Id == stageRun1Id {
			stageRun1Found = true
		} else if stageRun.Id == stageRun2Id {
			stageRun2Found = true
		}
	}

	if !stageRun1Found {
		test.Fatal(fmt.Sprintf("Failed to find stage run with id: %d", stageRun1Id))
	} else if !stageRun2Found {
		test.Fatal(fmt.Sprintf("Failed to find stage run with id: %d", stageRun2Id))
	}

	err = connection.Stages.Update.SetReturnCode(stageRun1Id, 17)
	if err != nil {
		test.Fatal(err)
	}

	stageRun, err := connection.Stages.Read.GetRun(stageRun1Id)
	if err != nil {
		test.Fatal(err)
	} else if stageRun.ReturnCode != 17 {
		test.Fatal("stageRun.ReturnCode mismatch")
	}

	err = connection.Stages.Update.SetEndTime(stageRun1Id, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Stages.Update.SetStartTime(stageRun1Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	err = connection.Stages.Update.SetStartTime(stageRun1Id, time.Now())
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.SetEndTime(stageRun1Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	err = connection.Stages.Update.SetEndTime(stageRun1Id, time.Now())
	if err != nil {
		test.Fatal(err)
	}
}

func TestConsoleText(test *testing.T) {
	stageId, err := connection.Stages.Create.Create(stageVerificationId, stageName+"2", stageFlavor, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRunId, err := connection.Stages.Create.CreateRun(stageId)
	if err != nil {
		test.Fatal(err)
	}

	lines := map[uint64]string{
		0:  "hello",
		7:  "there",
		42: "sir",
	}
	err = connection.Stages.Update.AddConsoleLines(stageRunId, lines)
	if err != nil {
		test.Fatal(err)
	}

	lines, err = connection.Stages.Read.GetAllConsoleText(stageRunId)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 3 {
		test.Fatal("Expected three lines of console text in result")
	}

	lines, err = connection.Stages.Read.GetConsoleTextHead(stageRunId, 7, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 1 {
		test.Fatal("Expected one line of console text in result")
	}

	lines, err = connection.Stages.Read.GetConsoleTextTail(stageRunId, 0, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 1 {
		test.Fatal("Expected one line of console text in result")
	}
}
