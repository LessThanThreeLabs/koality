package database

import (
	"fmt"
	"koality/resources"
	"sync"
	"time"
)

var (
	cleanTimerSync sync.Once
)

// TODO: write a test for this when it's implemented
func Clean(connection *resources.Connection) error {
	fmt.Println("Need to clean...")
	return nil
}

func KeepClean(connection *resources.Connection) {
	go cleanTimerSync.Do(createStartCleanTimer(connection))
}

func createStartCleanTimer(connection *resources.Connection) func() {
	startCleanTimer := func() {
		timer := time.NewTimer(getDurationUntilNextClean())
		for _ = range timer.C {
			err := Clean(connection)
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
