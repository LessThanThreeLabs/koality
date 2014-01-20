package database

import (
	"fmt"
	"koality/resources"
	"koality/resources/database/textsamples"
	"math/rand"
	"time"
)

const (
	numRepositories               = 2
	numVerificationsPerRepository = 10
	parallelizationLevel          = 2
	nullSnapshotId                = 0
)

var (
	dumpStaleTime time.Time = time.Now()
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func PopulateDatabase() error {
	err := makeSureDumpExists()
	if err != nil {
		return err
	}

	err = LoadDump()
	return err
}

func makeSureDumpExists() error {
	exists, err := DumpExistsAndNotStale(dumpStaleTime)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	err = Reseed()
	if err != nil {
		return err
	}

	connection, err := New()
	if err != nil {
		return err
	}

	err = createUsers(connection)
	if err != nil {
		return err
	}

	err = createRepositories(connection)
	if err != nil {
		return err
	}

	err = CreateDump()
	return err
}

func createUsers(connection *resources.Connection) error {
	userPasswordHash := []byte("password-hash")
	userPasswordSalt := []byte("password-salt")

	_, err := connection.Users.Create.Create("admin@koalitycode.com", "Mister", "Admin", userPasswordHash, userPasswordSalt, true)
	if err != nil {
		return err
	}

	_, err = connection.Users.Create.Create("jchu@koalitycode.com", "Jonathan", "Chu", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		return err
	}

	_, err = connection.Users.Create.Create("jpotter@koalitycode.com", "Jordan", "Potter", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		return err
	}

	_, err = connection.Users.Create.Create("bbland@koalitycode.com", "Brian", "Bland", userPasswordHash, userPasswordSalt, false)
	return err
}

func createRepositories(connection *resources.Connection) error {
	errorChannel := make(chan error, numRepositories)

	for index := 0; index < numRepositories; index++ {
		go func(index int) {
			repositoryName := fmt.Sprintf("Repository-%d", index)
			repositoryLocalUri := fmt.Sprintf("git/test-data.koalitycode.com/koality-%d.git", index)
			repositoryRemoteUri := fmt.Sprintf("git@github.com:KoalityCode/koality-%d.git", index)
			repository, err := connection.Repositories.Create.Create(repositoryName, "git", repositoryLocalUri, repositoryRemoteUri)
			if err != nil {
				errorChannel <- err
				return
			}

			err = createVerifications(connection, repository.Id, numVerificationsPerRepository)
			errorChannel <- err
		}(index)
	}

	for index := 0; index < numRepositories; index++ {
		err := <-errorChannel
		if err != nil {
			return err
		}
	}
	return nil
}

func createVerifications(connection *resources.Connection, repositoryId uint64, numVerifications int) error {
	userNames := []string{"Jonathan Chu", "Jordan Potter", "Brian Bland", "Andrey Kostov"}
	userEmails := []string{"jchu@koalitycode.com", "jpotter@koalitycode.com", "bbland@koalitycode.com", "akostov@koalitycode.com"}
	mergeTargets := []string{"master", "development", "feature_branch_1", "feature_branch_2"}

	createSha := func() string {
		shaChars := "0123456789ABCDEF"
		sha := ""
		for index := 0; index < 40; index++ {
			randomChar := shaChars[rand.Intn(len(shaChars))]
			sha = sha + string(randomChar)
		}
		return sha
	}

	errorChannel := make(chan error, numVerifications)

	for index := 0; index < numVerifications; index++ {
		go func(index int) {
			headMessage := fmt.Sprintf("This is a commit from %s", userNames[index%len(userNames)])
			verification, err := connection.Verifications.Create.Create(repositoryId, nullSnapshotId, createSha(), createSha(),
				headMessage, userNames[index%len(userNames)], userEmails[index%len(userEmails)],
				mergeTargets[index%len(mergeTargets)], userEmails[rand.Intn(len(userEmails))])
			if err != nil {
				errorChannel <- err
				return
			}

			err = createStages(connection, verification.Id)
			errorChannel <- err
		}(index)
	}

	for index := 0; index < numVerificationsPerRepository; index++ {
		err := <-errorChannel
		if err != nil {
			return err
		}
	}
	return nil
}

func createStages(connection *resources.Connection, verificationId uint64) error {
	stageNames := []string{"install dependencies", "prepare database", "frontend tests", "backend tests"}

	for index, stageName := range stageNames {
		stage, err := connection.Stages.Create.Create(verificationId, uint64(index/2), stageName, uint64(index%2))
		if err != nil {
			return err
		}

		err = createStageRuns(connection, stage.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func createStageRuns(connection *resources.Connection, stageId uint64) error {
	getReturnCode := func() int {
		if rand.Float32() > .2 {
			return 0
		} else {
			return rand.Intn(254) + 1
		}
	}

	for index := 0; index < parallelizationLevel; index++ {
		stageRun, err := connection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return err
		}

		err = connection.Stages.Update.SetStartTime(stageRun.Id, time.Now())
		if err != nil {
			return err
		}

		err = addConsoleText(connection, stageRun.Id)
		if err != nil {
			return err
		}

		err = connection.Stages.Update.SetReturnCode(stageRun.Id, getReturnCode())
		if err != nil {
			return err
		}

		err = connection.Stages.Update.SetEndTime(stageRun.Id, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}

func addConsoleText(connection *resources.Connection, stageRunId uint64) error {
	books := [][]string{textsamples.IRobot, textsamples.AliceInWonderland, textsamples.GreatExpectations}

	textToTextMap := func(text []string) map[uint64]string {
		textMap := make(map[uint64]string, len(text))
		for lineNumber, line := range text {
			textMap[uint64(lineNumber)] = line
		}
		return textMap
	}

	text := books[rand.Intn(len(books))]
	maxLines := rand.Intn(len(text)-750) + 200
	textToAdd := text[0:maxLines]
	err := connection.Stages.Update.AddConsoleLines(stageRunId, textToTextMap(textToAdd))
	return err
}
