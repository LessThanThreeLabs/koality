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
	stageSectionNumber        = 4
	stageName                 = "awesome stage"
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
	_, err := connection.Stages.Create.Create(0, stageSectionNumber, stageName, stageOrderNumber)
	if _, ok := err.(resources.NoSuchVerificationError); !ok {
		test.Fatal("Expected NoSuchVerificationError when providing invalid verification id")
	}
	_, err = connection.Stages.Create.Create(stageVerificationId, stageSectionNumber, "", stageOrderNumber)
	if err == nil {
		test.Fatal("Expected error after providing invalid stage name")
	}
}

func TestCreateStage(test *testing.T) {
	stageId, err := connection.Stages.Create.Create(stageVerificationId, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stage, err := connection.Stages.Read.Get(stageId)
	if err != nil {
		test.Fatal(err)
	} else if stage.Id != stageId {
		test.Fatal("stage.Id mismatch")
	}

	stage, err = connection.Stages.Read.GetBySectionNumberAndName(stage.VerificationId, stage.SectionNumber, stage.Name)
	if err != nil {
		test.Fatal(err)
	} else if stage == nil {
		test.Fatal("Unable to find stage")
	}

	stages, err := connection.Stages.Read.GetAll(stageVerificationId)
	if err != nil {
		test.Fatal(err)
	} else if len(stages) != 1 {
		test.Fatal("Unexpected number of stages")
	}

	_, err = connection.Stages.Create.Create(stageVerificationId, stageSectionNumber, stageName, stageOrderNumber)
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

	stageRuns, err := connection.Stages.Read.GetAllRuns(stageId)
	if err != nil {
		test.Fatal(err)
	} else if len(stageRuns) != 2 {
		test.Fatal("Unexpected number of stage runs")
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
	stageId, err := connection.Stages.Create.Create(stageVerificationId, stageSectionNumber, stageName+"-console-text", stageOrderNumber)
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

func TestXunit(test *testing.T) {
	stageId, err := connection.Stages.Create.Create(stageVerificationId, stageSectionNumber, stageName+"-xunit", stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRunId, err := connection.Stages.Create.CreateRun(stageId)
	if err != nil {
		test.Fatal(err)
	}

	firstXunitResult := resources.XunitResult{"first", "some/path/1.xml", "", "", "", "", time.Now(), 1.137}
	secondXunitResult := resources.XunitResult{"second", "some/path/2.xml", "", "", "", "", time.Now(), 1.137}
	xunitResults := []resources.XunitResult{firstXunitResult, secondXunitResult}
	err = connection.Stages.Update.AddXunitResults(stageRunId, xunitResults)
	if err != nil {
		test.Fatal(err)
	}

	returnedXunitResults, err := connection.Stages.Read.GetXunitResults(stageRunId)
	if err != nil {
		test.Fatal(err)
	}

	if len(returnedXunitResults) != 2 {
		test.Fatal("Unexpected numebr of xunit results")
	}
}

func TestExport(test *testing.T) {
	stageId, err := connection.Stages.Create.Create(stageVerificationId, stageSectionNumber, stageName+"-export", stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRunId, err := connection.Stages.Create.CreateRun(stageId)
	if err != nil {
		test.Fatal(err)
	}

	firstExport := resources.Export{"some/path/1.xml", "https://s3.aws/bucket-name/1.xml"}
	secondExport := resources.Export{"some/path/2.xml", "https://s3.aws/bucket-name/2.xml"}
	exports := []resources.Export{firstExport, secondExport}
	err = connection.Stages.Update.AddExports(stageRunId, exports)
	if err != nil {
		test.Fatal(err)
	}

	returnedExports, err := connection.Stages.Read.GetExports(stageRunId)
	if err != nil {
		test.Fatal(err)
	}

	if len(returnedExports) != 2 {
		test.Fatal("Unexpecetd number of exports")
	}
}
