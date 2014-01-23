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
	numPools                      = 1
	parallelizationLevel          = 2
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

	if err = Reseed(); err != nil {
		return err
	}

	connection, err := New()
	if err != nil {
		return err
	}

	if err = createUsers(connection); err != nil {
		return err
	}

	if err = createRepositories(connection); err != nil {
		return err
	}

	if err = createPools(connection); err != nil {
		return err
	}

	return CreateDump()
}

func createUsers(connection *resources.Connection) error {
	publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxvvK4FBlsGzexbr" +
		"5IMEfvp0LfaPg2LHJlrHPqawe66136PrXPQHDJUN5rUb8LEulVVMsW6fRjG5oAy" +
		"tmOZ/DCGlxLN7vN65c8adw67lLjHVpQ8uHJteRkq0EuL/rZSPBLm2fP/yAeJYRi" +
		"JP6fob24PpklwIz5cr9tGHH7DJmzk69PzU3AdL7DbUZ/vIay9cPFV5sQ3BGTpHS" +
		"QlKunWWtN+m6Lp5ZAwY6+bvdw9E/8PYp7+aBRpbPDJ4f3uiMzcmzSxPqcoz+PuC" +
		"zljHeYmm/vYF2XmeB66cAzPSig3xAz5YVgTFBW9FWvg6W5DcdPsUQGqeyJta7pp" +
		"IQW88HOpNk5"

	passwordHasher, err := resources.NewPasswordHasher()
	if err != nil {
		return err
	}

	createUser := func(email, firstName, lastName, password string, admin bool) error {
		passwordHash, passwordSalt, err := passwordHasher.GenerateHashAndSalt(password)
		if err != nil {
			return err
		}

		user, err := connection.Users.Create.Create(email, firstName, lastName, passwordHash, passwordSalt, admin)
		if err != nil {
			return err
		}

		_, err = connection.Users.Update.AddKey(user.Id, "my key", publicKey+" "+user.Email)
		return err
	}

	err = createUser("admin@koalitycode.com", "Mister", "Admin", "admin123", true)
	if err != nil {
		return err
	}

	err = createUser("jchu@koalitycode.com", "Jonathan", "Chu", "jchu123", false)
	if err != nil {
		return err
	}

	err = createUser("jpotter@koalitycode.com", "Jordan", "Potter", "jpotter123", false)
	if err != nil {
		return err
	}

	err = createUser("bbland@koalitycode.com", "Brian", "Bland", "bbland123", false)
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
			verification, err := connection.Verifications.Create.Create(repositoryId, createSha(), createSha(),
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

func createPools(connection *resources.Connection) error {
	accessKey := "aaaabbbbccccddddeeee"
	secretKey := "0000111122223333444455556666777788889999"
	username := "koality"
	baseAmiId := "ami-12345678"
	securityGroupId := "sg-12345678"
	vpcSubnetId := "subnet-12345678"
	instanceType := "m1.medium"
	numReadyInstances := uint64(2)
	numMaxInstances := uint64(10)
	rootDriveSize := uint64(100)
	userData := "echo hello"

	for index := 0; index < numPools; index++ {
		name := fmt.Sprintf("Pool-%d", index)
		_, err := connection.Pools.Create.CreateEc2Pool(name, accessKey, secretKey, username, baseAmiId, securityGroupId, vpcSubnetId,
			instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData)
		if err != nil {
			return err
		}
	}
	return nil
}
