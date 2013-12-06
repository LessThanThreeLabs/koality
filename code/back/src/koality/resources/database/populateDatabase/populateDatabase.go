package main

import (
	"fmt"
	"koality/resources"
	"koality/resources/database"
	"math/rand"
	"time"
)

const (
	numRepositories      = 3
	numVerifications     = 100
	parallelizationLevel = 4
)

func main() {
	fmt.Printf("Reseeding database (%s)\n", getTimeString())
	database.Reseed()

	connection, err := database.New()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Beginning to populate database (%s)\n", getTimeString())
	err = createUsers(connection)
	if err != nil {
		panic(err)
	}

	repositoryIds, err := createRepositories(connection)
	if err != nil {
		panic(err)
	}

	verificationIds, err := createVerifications(connection, repositoryIds)
	if err != nil {
		panic(err)
	}

	stageIds, err := createStages(connection, verificationIds)
	if err != nil {
		panic(err)
	}

	stageRunIds, err := createStageRuns(connection, stageIds)
	if err != nil {
		panic(err)
	}

	err = addConsoleText(connection, stageRunIds)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Completed populating database (%s)\n", getTimeString())

	fmt.Printf("TODO: Saving dump (%s)\n", getTimeString())
}

func getTimeString() string {
	currentTime := time.Now()
	return fmt.Sprintf("%d:%02d:%d.%d", currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond())
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
	if err != nil {
		return err
	}

	return nil
}

func createRepositories(connection *resources.Connection) ([]uint64, error) {
	repositoryIds := make([]uint64, 0, numRepositories)

	for index := 0; index < numRepositories; index++ {
		repositoryName := fmt.Sprintf("Repository %d", index)
		repositoryLocalUri := fmt.Sprintf("git@test-data.koalitycode.com:koality-%d.git", index)
		repositoryRemoteUri := fmt.Sprintf("git@github.com:KoalityCode/koality-%d.git", index)
		repositoryId, err := connection.Repositories.Create.Create(repositoryName, "git", repositoryLocalUri, repositoryRemoteUri)
		if err != nil {
			return nil, err
		}
		repositoryIds = append(repositoryIds, repositoryId)
	}

	return repositoryIds, nil
}

func createVerifications(connection *resources.Connection, repositoryIds []uint64) ([]uint64, error) {
	verificationIds := make([]uint64, 0, numVerifications)

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

	for _, repositoryId := range repositoryIds {
		numVerificationsToCreate := numVerifications / numRepositories
		if repositoryId == repositoryIds[len(repositoryIds)-1] {
			numVerificationsToCreate = numVerifications/numRepositories + numVerifications%numRepositories
		}

		for index := 0; index < numVerificationsToCreate; index++ {
			headMessage := fmt.Sprintf("This is a commit from %s", userNames[index%len(userNames)])
			verificationId, err := connection.Verifications.Create.Create(repositoryId, createSha(), createSha(),
				headMessage, userNames[index%len(userNames)], userEmails[index%len(userEmails)],
				mergeTargets[index%len(mergeTargets)], userEmails[rand.Intn(len(userEmails))])
			if err != nil {
				return nil, err
			}
			verificationIds = append(verificationIds, verificationId)
		}
	}

	return verificationIds, nil
}

func createStages(connection *resources.Connection, verificationIds []uint64) ([]uint64, error) {
	stageNames := []string{"install dependencies", "prepare database", "frontend tests", "backend tests"}
	stageIds := make([]uint64, 0, numVerifications*len(stageNames))

	for _, verificationId := range verificationIds {
		for index, stageName := range stageNames {
			stageId, err := connection.Stages.Create.Create(verificationId, uint64(index/2), stageName, uint64(index%2))
			if err != nil {
				return nil, err
			}
			stageIds = append(stageIds, stageId)
		}
	}

	return stageIds, nil
}

func createStageRuns(connection *resources.Connection, stageIds []uint64) ([]uint64, error) {
	stageRunIds := make([]uint64, 0, len(stageIds)*parallelizationLevel)

	getReturnCode := func() int {
		if rand.Float32() > .2 {
			return 0
		} else {
			return rand.Intn(254) + 1
		}
	}

	for _, stageId := range stageIds {
		stageRunId, err := connection.Stages.Create.CreateRun(stageId)
		if err != nil {
			return nil, err
		}

		err = connection.Stages.Update.SetStartTime(stageRunId, time.Now())
		if err != nil {
			return nil, err
		}

		err = connection.Stages.Update.SetReturnCode(stageRunId, getReturnCode())
		if err != nil {
			return nil, err
		}

		err = connection.Stages.Update.SetEndTime(stageRunId, time.Now())
		if err != nil {
			return nil, err
		}

		stageRunIds = append(stageRunIds, stageRunId)
	}

	return stageRunIds, nil
}

func addConsoleText(connection *resources.Connection, stageRunIds []uint64) error {
	books := [][]string{iRobot, aliceInWonderland, greatExpectations}

	textToTextMap := func(text []string) map[uint64]string {
		textMap := make(map[uint64]string, len(text))
		for lineNumber, line := range text {
			textMap[uint64(lineNumber)] = line
		}
		return textMap
	}

	for _, stageRunId := range stageRunIds {
		text := books[rand.Intn(len(books))]
		maxLines := rand.Intn(len(text)-200) + 200
		textToAdd := text[0:maxLines]
		err := connection.Stages.Update.AddConsoleLines(stageRunId, textToTextMap(textToAdd))
		if err != nil {
			return err
		}
	}

	return nil
}
