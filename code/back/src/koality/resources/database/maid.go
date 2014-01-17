package database

import (
	"koality/resources"
	"math"
	"sync"
	"time"
)

const (
	defaultNumVerificationsToRetain = 250
)

var (
	cleanTimerSync sync.Once
)

func Clean(connection *resources.Connection, numVerificationsToRetain uint32) error {
	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		return err
	}

	verificationCount := 0
	errorChannel := make(chan error, 100)
	for _, repository := range repositories {
		oldVerifications, err := connection.Verifications.Read.GetTail(repository.Id, numVerificationsToRetain, math.MaxUint32)
		if err != nil {
			return err
		}

		for _, oldVerification := range oldVerifications {
			verificationCount++
			go func(verificationId uint64) {
				err := cleanVerification(connection, verificationId)
				errorChannel <- err
			}(oldVerification.Id)
		}
	}

	for index := 0; index < verificationCount; index++ {
		err := <-errorChannel
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanVerification(connection *resources.Connection, verificationId uint64) error {
	stages, err := connection.Stages.Read.GetAll(verificationId)
	if err != nil {
		return err
	}

	for _, stage := range stages {
		for _, stageRun := range stage.Runs {
			err = connection.Stages.Update.RemoveAllConsoleLines(stageRun.Id)
			if err != nil {
				return err
			}

			err = connection.Stages.Update.RemoveAllXunitResults(stageRun.Id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func KeepClean(connection *resources.Connection) {
	go cleanTimerSync.Do(createStartCleanTimer(connection))
}

func createStartCleanTimer(connection *resources.Connection) func() {
	startCleanTimer := func() {
		timer := time.NewTimer(getDurationUntilNextClean())
		for _ = range timer.C {
			err := Clean(connection, defaultNumVerificationsToRetain)
			if err != nil {
				panic(err)
			}
			timer.Reset(getDurationUntilNextClean())
		}
	}

	return startCleanTimer
}

func getDurationUntilNextClean() time.Duration {
	currentTime := time.Now()
	location, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}

	// TODO: need to pull the preferred cleanup time from the database
	// 3:00am PDT
	nextTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()+1, 11, 0, 0, 0, location)
	return nextTime.Sub(currentTime)
}
