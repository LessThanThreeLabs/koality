package main

import (
	"fmt"
	"koality/resources"
	"koality/resources/database"
	"math/rand"
	"time"
)

const (
	numRepositories               = 3
	numVerificationsPerRepository = 25
	parallelizationLevel          = 2
)

func main() {
	fmt.Printf("Reseeding database (%s)\n", getTimeString())
	database.Reseed()

	connection, err := database.New()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Beginning to populate database (%s)\n", getTimeString())
	createUsers(connection)
	createRepositories(connection)
	fmt.Printf("Completed populating database (%s)\n", getTimeString())

	fmt.Printf("Saving dump (%s)\n", getTimeString())
	err = database.SaveDump()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Completed saving dump (%s)\n", getTimeString())
}

func getTimeString() string {
	currentTime := time.Now()
	return fmt.Sprintf("%d:%02d:%d.%d", currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond())
}

func createUsers(connection *resources.Connection) {
	userPasswordHash := []byte("password-hash")
	userPasswordSalt := []byte("password-salt")

	_, err := connection.Users.Create.Create("admin@koalitycode.com", "Mister", "Admin", userPasswordHash, userPasswordSalt, true)
	if err != nil {
		panic(err)
	}

	_, err = connection.Users.Create.Create("jchu@koalitycode.com", "Jonathan", "Chu", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		panic(err)
	}

	_, err = connection.Users.Create.Create("jpotter@koalitycode.com", "Jordan", "Potter", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		panic(err)
	}

	_, err = connection.Users.Create.Create("bbland@koalitycode.com", "Brian", "Bland", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		panic(err)
	}
}

func createRepositories(connection *resources.Connection) {
	completedChannel := make(chan bool, numRepositories)

	for index := 0; index < numRepositories; index++ {
		go func(index int) {
			repositoryName := fmt.Sprintf("Repository %d", index)
			repositoryLocalUri := fmt.Sprintf("git@test-data.koalitycode.com:koality-%d.git", index)
			repositoryRemoteUri := fmt.Sprintf("git@github.com:KoalityCode/koality-%d.git", index)
			repositoryId, err := connection.Repositories.Create.Create(repositoryName, "git", repositoryLocalUri, repositoryRemoteUri)
			if err != nil {
				panic(err)
			}

			createVerifications(connection, repositoryId)
			completedChannel <- true
		}(index)
	}

	for index := 0; index < numRepositories; index++ {
		<-completedChannel
	}
}

func createVerifications(connection *resources.Connection, repositoryId uint64) {
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

	completedChannel := make(chan bool, numVerificationsPerRepository)

	for index := 0; index < numVerificationsPerRepository; index++ {
		go func(index int) {
			headMessage := fmt.Sprintf("This is a commit from %s", userNames[index%len(userNames)])
			verificationId, err := connection.Verifications.Create.Create(repositoryId, createSha(), createSha(),
				headMessage, userNames[index%len(userNames)], userEmails[index%len(userEmails)],
				mergeTargets[index%len(mergeTargets)], userEmails[rand.Intn(len(userEmails))])
			if err != nil {
				panic(err)
			}

			createStages(connection, verificationId)
			completedChannel <- true
		}(index)
	}

	for index := 0; index < numVerificationsPerRepository; index++ {
		<-completedChannel
	}
}

func createStages(connection *resources.Connection, verificationId uint64) {
	stageNames := []string{"install dependencies", "prepare database", "frontend tests", "backend tests"}

	for index, stageName := range stageNames {
		stageId, err := connection.Stages.Create.Create(verificationId, uint64(index/2), stageName, uint64(index%2))
		if err != nil {
			panic(err)
		}

		createStageRuns(connection, stageId)
	}
}

func createStageRuns(connection *resources.Connection, stageId uint64) {
	getReturnCode := func() int {
		if rand.Float32() > .2 {
			return 0
		} else {
			return rand.Intn(254) + 1
		}
	}

	for index := 0; index < parallelizationLevel; index++ {
		stageRunId, err := connection.Stages.Create.CreateRun(stageId)
		if err != nil {
			panic(err)
		}

		err = connection.Stages.Update.SetStartTime(stageRunId, time.Now())
		if err != nil {
			panic(err)
		}

		addConsoleText(connection, stageRunId)

		err = connection.Stages.Update.SetReturnCode(stageRunId, getReturnCode())
		if err != nil {
			panic(err)
		}

		err = connection.Stages.Update.SetEndTime(stageRunId, time.Now())
		if err != nil {
			panic(err)
		}
	}
}

func addConsoleText(connection *resources.Connection, stageRunId uint64) {
	books := [][]string{iRobot, aliceInWonderland, greatExpectations}

	textToTextMap := func(text []string) map[uint64]string {
		textMap := make(map[uint64]string, len(text))
		for lineNumber, line := range text {
			textMap[uint64(lineNumber)] = line
		}
		return textMap
	}

	text := books[rand.Intn(len(books))]
	maxLines := rand.Intn(len(text)-200) + 200
	textToAdd := text[0:maxLines]
	err := connection.Stages.Update.AddConsoleLines(stageRunId, textToTextMap(textToAdd))
	if err != nil {
		panic(err)
	}
}
