package database

import (
	"fmt"
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidStage(test *testing.T) {
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
	}
	firstVerification := verifications[0]

	stageSectionNumber := uint64(4)
	stageName := "awesome stage"
	stageOrderNumber := uint64(17)

	_, err = connection.Stages.Create.Create(0, stageSectionNumber, stageName, stageOrderNumber)
	if _, ok := err.(resources.NoSuchVerificationError); !ok {
		test.Fatal("Expected NoSuchVerificationError when providing invalid verification id")
	}
	_, err = connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, "", stageOrderNumber)
	if err == nil {
		test.Fatal("Expected error after providing invalid stage name")
	}
}

func TestCreateStage(test *testing.T) {
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
	}
	firstVerification := verifications[0]

	stageSectionNumber := uint64(4)
	stageName := "awesome stage"
	stageOrderNumber := uint64(17)

	stageId, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
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

	stages, err := connection.Stages.Read.GetAll(firstVerification.Id)
	if err != nil {
		test.Fatal(err)
	} else if len(stages) == 0 {
		test.Fatal("Unexpected number of stages")
	}

	_, err = connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
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

	err = connection.Stages.Update.SetReturnCode(0, 17)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to set return code for nonexistent stage")
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
}

func TestStageTimes(test *testing.T) {
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
	}
	firstVerification := verifications[0]

	stageSectionNumber := uint64(4)
	stageName := "awesome stage"
	stageOrderNumber := uint64(17)

	stageId, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRunId, err := connection.Stages.Create.CreateRun(stageId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.SetEndTime(stageRunId, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Stages.Update.SetStartTime(stageRunId, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	err = connection.Stages.Update.SetStartTime(stageRunId, time.Now())
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.SetStartTime(0, time.Now())
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to set start time for nonexistent stage run")
	}

	err = connection.Stages.Update.SetEndTime(stageRunId, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	err = connection.Stages.Update.SetEndTime(stageRunId, time.Now())
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.SetEndTime(0, time.Now())
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to set end time for nonexistent stage run")
	}
}

func TestConsoleLines(test *testing.T) {
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
	}
	firstVerification := verifications[0]

	stageSectionNumber := uint64(4)
	stageName := "awesome stage"
	stageOrderNumber := uint64(17)

	stageId, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
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

	// The existence of a stage run id isn't checked
	// when adding console lines for performance reasons
	// err = connection.Stages.Update.AddConsoleLines(0, lines)
	// if _, ok := err.(resources.NoSuchStageRunError); !ok {
	// 	test.Fatal("Expected NoSuchStageRunError when trying to add console lines for nonexistent stage run")
	// }

	lines, err = connection.Stages.Read.GetAllConsoleLines(stageRunId)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 3 {
		test.Fatal("Expected three lines of console in result")
	}

	lines, err = connection.Stages.Read.GetConsoleLinesHead(stageRunId, 7, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 1 {
		test.Fatal("Expected one line of console in result")
	}

	lines, err = connection.Stages.Read.GetConsoleLinesTail(stageRunId, 0, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 1 {
		test.Fatal("Expected one line of console in result")
	}

	err = connection.Stages.Update.RemoveAllConsoleLines(stageRunId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.RemoveAllConsoleLines(0)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to delete console lines for nonexistent stage run")
	}

	lines, err = connection.Stages.Read.GetAllConsoleLines(stageRunId)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 0 {
		test.Fatal("Expected zero lines of console in result")
	}
}

func TestXunit(test *testing.T) {
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
	}
	firstVerification := verifications[0]

	stageSectionNumber := uint64(4)
	stageName := "awesome stage"
	stageOrderNumber := uint64(17)

	stageId, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
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

	err = connection.Stages.Update.AddXunitResults(0, xunitResults)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to add xunit results for nonexistent stage run")
	}

	returnedXunitResults, err := connection.Stages.Read.GetAllXunitResults(stageRunId)
	if err != nil {
		test.Fatal(err)
	} else if len(returnedXunitResults) != 2 {
		test.Fatal("Expected two xunit results")
	}

	err = connection.Stages.Update.RemoveAllXunitResults(stageRunId)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.RemoveAllXunitResults(0)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to delete xunit results for nonexistent stage run")
	}

	returnedXunitResults, err = connection.Stages.Read.GetAllXunitResults(stageRunId)
	if err != nil {
		test.Fatal(err)
	} else if len(returnedXunitResults) != 0 {
		test.Fatal("Expected zero xunit results")
	}
}

func TestExport(test *testing.T) {
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
	}
	firstVerification := verifications[0]

	stageSectionNumber := uint64(4)
	stageName := "awesome stage"
	stageOrderNumber := uint64(17)

	stageId, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
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
