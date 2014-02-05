package database

import (
	"koality/resources"
	"math"
	"sync"
	"time"
)

const (
	defaultNumBuildsToRetain = 250
)

var (
	cleanTimerSync sync.Once
)

func Clean(connection *resources.Connection, numBuildsToRetain uint32) error {
	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		return err
	}

	buildCount := 0
	errorChannel := make(chan error, 100)
	for _, repository := range repositories {
		oldBuilds, err := connection.Builds.Read.GetTail(repository.Id, numBuildsToRetain, math.MaxUint32)
		if err != nil {
			return err
		}

		for _, oldBuild := range oldBuilds {
			buildCount++
			go func(buildId uint64) {
				err := cleanBuild(connection, buildId)
				errorChannel <- err
			}(oldBuild.Id)
		}
	}

	for index := 0; index < buildCount; index++ {
		err := <-errorChannel
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanBuild(connection *resources.Connection, buildId uint64) error {
	stages, err := connection.Stages.Read.GetAll(buildId)
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
			err := Clean(connection, defaultNumBuildsToRetain)
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
