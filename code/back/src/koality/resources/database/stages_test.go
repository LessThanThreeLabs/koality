package database

import (
	"fmt"
	"koality/resources"
	"testing"
	"time"
)

func TestCreateInvalidStage(test *testing.T) {
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
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	createEventReceived := make(chan bool, 1)
	var createEventStage *resources.Stage
	stageCreatedHandler := func(stage *resources.Stage) {
		createEventStage = stage
		createEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToCreatedEvents(stageCreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	stageRunCreatedEventReceived := make(chan bool, 1)
	var stageRunCreatedEventStageRun *resources.StageRun
	stageRunCreatedHandler := func(stageRun *resources.StageRun) {
		stageRunCreatedEventStageRun = stageRun
		stageRunCreatedEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToRunCreatedEvents(stageRunCreatedHandler)
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

	stage, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	if stage.SectionNumber != stageSectionNumber {
		test.Fatal("stage.SectionNumber mismatch")
	} else if stage.Name != stageName {
		test.Fatal("stage.Name mismatch")
	} else if stage.OrderNumber != stageOrderNumber {
		test.Fatal("stage.OrderNumber mismatch")
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-createEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage creation event")
	}

	if createEventStage.Id != stage.Id {
		test.Fatal("Bad stage.Id in stage creation event")
	} else if createEventStage.SectionNumber != stage.SectionNumber {
		test.Fatal("Bad stage.SectionNumber in stage creation event")
	} else if createEventStage.Name != stage.Name {
		test.Fatal("Bad stage.Name in stage creation event")
	} else if createEventStage.OrderNumber != stage.OrderNumber {
		test.Fatal("Bad stage.OrderNumber in stage creation event")
	}

	stage2, err := connection.Stages.Read.Get(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	if stage.Id != stage2.Id {
		test.Fatal("stage.Id mismatch")
	} else if stage.SectionNumber != stage2.SectionNumber {
		test.Fatal("stage.SectionNumber mismatch")
	} else if stage.Name != stage2.Name {
		test.Fatal("stage.Name mismatch")
	} else if stage.OrderNumber != stage2.OrderNumber {
		test.Fatal("stage.OrderNumber mismatch")
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

	stageRun1, err := connection.Stages.Create.CreateRun(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-stageRunCreatedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run creation event")
	}

	if stageRunCreatedEventStageRun.Id != stageRun1.Id {
		test.Fatal("Bad stageRun.Id in stage run creation event")
	}

	stageRun2, err := connection.Stages.Create.CreateRun(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-stageRunCreatedEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run creation event")
	}

	if stageRunCreatedEventStageRun.Id != stageRun2.Id {
		test.Fatal("Bad stageRun.Id in stage run creation event")
	}

	stage, err = connection.Stages.Read.Get(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	stageRun1Found, stageRun2Found := false, false
	for _, stageRun := range stage.Runs {
		if stageRun.Id == stageRun1.Id {
			stageRun1Found = true
		} else if stageRun.Id == stageRun2.Id {
			stageRun2Found = true
		}
	}

	if !stageRun1Found {
		test.Fatal(fmt.Sprintf("Failed to find stage run with id: %d", stageRun1.Id))
	} else if !stageRun2Found {
		test.Fatal(fmt.Sprintf("Failed to find stage run with id: %d", stageRun2.Id))
	}

	stageRuns, err := connection.Stages.Read.GetAllRuns(stage.Id)
	if err != nil {
		test.Fatal(err)
	} else if len(stageRuns) != 2 {
		test.Fatal("Unexpected number of stage runs")
	}
}

func TestStageReturnCode(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	stageRunEventReceived := make(chan bool, 1)
	stageRunEventId := uint64(0)
	stageRunEventReturnCode := 0
	stageRunHandler := func(stageId uint64, returnCode int) {
		stageRunEventId = stageId
		stageRunEventReturnCode = returnCode
		stageRunEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToReturnCodeUpdatedEvents(stageRunHandler)
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

	stage, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRun, err := connection.Stages.Create.CreateRun(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	returnCode := 17
	err = connection.Stages.Update.SetReturnCode(stageRun.Id, returnCode)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-stageRunEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run return code updated event")
	}

	if stageRunEventId != stageRun.Id {
		test.Fatal("Bad stageRun.Id in stage run return code updated event")
	} else if stageRunEventReturnCode != returnCode {
		test.Fatal("Bad return code in stage run return code updated event")
	}

	err = connection.Stages.Update.SetReturnCode(0, returnCode)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to set return code for nonexistent stage")
	}

	stageRun2, err := connection.Stages.Read.GetRun(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	} else if stageRun2.ReturnCode != returnCode {
		test.Fatal("stageRun.ReturnCode mismatch")
	}
}

func TestStageTimes(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	stageRunStartTimeEventReceived := make(chan bool, 1)
	stageRunStartTimeEventId := uint64(0)
	stageRunStartTimeEventTime := time.Now()
	stageRunStartTimeUpdatedHandler := func(stageRunId uint64, startTime time.Time) {
		stageRunStartTimeEventId = stageRunId
		stageRunStartTimeEventTime = startTime
		stageRunStartTimeEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToStartTimeUpdatedEvents(stageRunStartTimeUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	stageRunEndTimeEventReceived := make(chan bool, 1)
	stageRunEndTimeEventId := uint64(0)
	stageRunEndTimeEventTime := time.Now()
	stageRunEndTimeUpdatedHandler := func(stageRunId uint64, endTime time.Time) {
		stageRunEndTimeEventId = stageRunId
		stageRunEndTimeEventTime = endTime
		stageRunEndTimeEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToEndTimeUpdatedEvents(stageRunEndTimeUpdatedHandler)
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

	stage, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRun, err := connection.Stages.Create.CreateRun(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.SetEndTime(stageRun.Id, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Stages.Update.SetStartTime(stageRun.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	startTime := time.Now()
	err = connection.Stages.Update.SetStartTime(stageRun.Id, startTime)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-stageRunStartTimeEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run start time event")
	}

	if stageRunStartTimeEventId != stageRun.Id {
		test.Fatal("Bad stageRun.Id in start time event")
	} else if stageRunStartTimeEventTime != startTime {
		test.Fatal("Bad stage run start time in start time event")
	}

	err = connection.Stages.Update.SetStartTime(0, time.Now())
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to set start time for nonexistent stage run")
	}

	err = connection.Stages.Update.SetEndTime(stageRun.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	endTime := time.Now()
	err = connection.Stages.Update.SetEndTime(stageRun.Id, endTime)
	if err != nil {
		test.Fatal(err)
	}

	timeout = time.After(10 * time.Second)
	select {
	case <-stageRunEndTimeEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run end time event")
	}

	if stageRunEndTimeEventId != stageRun.Id {
		test.Fatal("Bad stageRun.Id in end time event")
	} else if stageRunEndTimeEventTime != endTime {
		test.Fatal("Bad stage run end time in end time event")
	}

	err = connection.Stages.Update.SetEndTime(0, time.Now())
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to set end time for nonexistent stage run")
	}
}

func TestConsoleLines(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	stageRunEventReceived := make(chan bool, 1)
	stageRunEventId := uint64(0)
	stageRunEventConsoleLines := map[uint64]string{}
	stageRunConsoleLinesAddedHandler := func(stageId uint64, consoleLines map[uint64]string) {
		stageRunEventId = stageId
		stageRunEventConsoleLines = consoleLines
		stageRunEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToConsoleLinesAddedEvents(stageRunConsoleLinesAddedHandler)
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

	stage, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRun, err := connection.Stages.Create.CreateRun(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	lines := map[uint64]string{
		0:  "hello",
		7:  "there",
		42: "sir",
	}
	err = connection.Stages.Update.AddConsoleLines(stageRun.Id, lines)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-stageRunEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run console lines added event")
	}

	if stageRunEventId != stageRun.Id {
		test.Fatal("Bad stageRun.Id in console lines added event")
	} else if len(stageRunEventConsoleLines) != len(lines) {
		test.Fatal("Bad console lines in console lines added event")
	} else if stageRunEventConsoleLines[0] != lines[0] {
		test.Fatal("Bad console lines in console lines added event")
	} else if stageRunEventConsoleLines[7] != lines[7] {
		test.Fatal("Bad console lines in console lines added event")
	} else if stageRunEventConsoleLines[42] != lines[42] {
		test.Fatal("Bad console lines in console lines added event")
	}

	// The existence of a stage run id isn't checked
	// when adding console lines for performance reasons
	// err = connection.Stages.Update.AddConsoleLines(0, lines)
	// if _, ok := err.(resources.NoSuchStageRunError); !ok {
	// 	test.Fatal("Expected NoSuchStageRunError when trying to add console lines for nonexistent stage run")
	// }

	lines, err = connection.Stages.Read.GetAllConsoleLines(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 3 {
		test.Fatal("Expected three lines of console in result")
	}

	lines, err = connection.Stages.Read.GetConsoleLinesHead(stageRun.Id, 7, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 1 {
		test.Fatal("Expected one line of console in result")
	}

	lines, err = connection.Stages.Read.GetConsoleLinesTail(stageRun.Id, 0, 1)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 1 {
		test.Fatal("Expected one line of console in result")
	}

	err = connection.Stages.Update.RemoveAllConsoleLines(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.RemoveAllConsoleLines(0)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to delete console lines for nonexistent stage run")
	}

	lines, err = connection.Stages.Read.GetAllConsoleLines(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	} else if len(lines) != 0 {
		test.Fatal("Expected zero lines of console in result")
	}
}

func TestXunit(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	stageRunEventReceived := make(chan bool, 1)
	stageRunEventId := uint64(0)
	stageRunEventXunitResults := []resources.XunitResult{}
	stageRunXunitResultsAddedHandler := func(stageId uint64, xunitResults []resources.XunitResult) {
		stageRunEventId = stageId
		stageRunEventXunitResults = xunitResults
		stageRunEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToXunitResultsAddedEvents(stageRunXunitResultsAddedHandler)
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

	stage, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRun, err := connection.Stages.Create.CreateRun(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	firstXunitResult := resources.XunitResult{
		Name:    "first",
		Path:    "some/path/1.xml",
		Seconds: 1.137,
	}
	secondXunitResult := resources.XunitResult{
		Name:    "second",
		Path:    "some/path/2.xml",
		Seconds: 1.137,
	}
	xunitResults := []resources.XunitResult{firstXunitResult, secondXunitResult}
	err = connection.Stages.Update.AddXunitResults(stageRun.Id, xunitResults)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-stageRunEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run xunit results added event")
	}

	if stageRunEventId != stageRun.Id {
		test.Fatal("Bad stageRun.Id in xunit results added event")
	} else if len(stageRunEventXunitResults) != len(xunitResults) {
		test.Fatal("Bad xunit results in xunit results added event")
	} else if stageRunEventXunitResults[0] != xunitResults[0] {
		test.Fatal("Bad xunit results in xunit results added event")
	} else if stageRunEventXunitResults[1] != xunitResults[1] {
		test.Fatal("Bad xunit results in xunit results added event")
	}

	err = connection.Stages.Update.AddXunitResults(0, xunitResults)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to add xunit results for nonexistent stage run")
	}

	returnedXunitResults, err := connection.Stages.Read.GetAllXunitResults(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	} else if len(returnedXunitResults) != 2 {
		test.Fatal("Expected two xunit results")
	}

	err = connection.Stages.Update.RemoveAllXunitResults(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Stages.Update.RemoveAllXunitResults(0)
	if _, ok := err.(resources.NoSuchStageRunError); !ok {
		test.Fatal("Expected NoSuchStageRunError when trying to delete xunit results for nonexistent stage run")
	}

	returnedXunitResults, err = connection.Stages.Read.GetAllXunitResults(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	} else if len(returnedXunitResults) != 0 {
		test.Fatal("Expected zero xunit results")
	}
}

func TestExport(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	stageRunEventReceived := make(chan bool, 1)
	stageRunEventId := uint64(0)
	stageRunEventExports := []resources.Export{}
	stageRunExportsAddedHandler := func(stageId uint64, exports []resources.Export) {
		stageRunEventId = stageId
		stageRunEventExports = exports
		stageRunEventReceived <- true
	}
	_, err = connection.Stages.Subscription.SubscribeToExportsAddedEvents(stageRunExportsAddedHandler)
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

	stage, err := connection.Stages.Create.Create(firstVerification.Id, stageSectionNumber, stageName, stageOrderNumber)
	if err != nil {
		test.Fatal(err)
	}

	stageRun, err := connection.Stages.Create.CreateRun(stage.Id)
	if err != nil {
		test.Fatal(err)
	}

	firstExport := resources.Export{"some/path/1.xml", "https://s3.aws/bucket-name/1.xml"}
	secondExport := resources.Export{"some/path/2.xml", "https://s3.aws/bucket-name/2.xml"}
	exports := []resources.Export{firstExport, secondExport}
	err = connection.Stages.Update.AddExports(stageRun.Id, exports)
	if err != nil {
		test.Fatal(err)
	}

	timeout := time.After(10 * time.Second)
	select {
	case <-stageRunEventReceived:
	case <-timeout:
		test.Fatal("Failed to hear stage run exports added event")
	}

	if stageRunEventId != stageRun.Id {
		test.Fatal("Bad stageRun.Id in expots added event")
	} else if len(stageRunEventExports) != len(exports) {
		test.Fatal("Bad exports in expots added event")
	} else if stageRunEventExports[0] != exports[0] {
		test.Fatal("Bad exports in expots added event")
	} else if stageRunEventExports[1] != exports[1] {
		test.Fatal("Bad exports in expots added event")
	}

	returnedExports, err := connection.Stages.Read.GetExports(stageRun.Id)
	if err != nil {
		test.Fatal(err)
	}

	if len(returnedExports) != 2 {
		test.Fatal("Unexpecetd number of exports")
	}
}
