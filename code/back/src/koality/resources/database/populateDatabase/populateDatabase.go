package main

import (
	"fmt"
	"koality/resources/database"
	"time"
)

func main() {
	fmt.Printf("Populating database... (%s)\n", getTimeString())

	err := database.PopulateDatabase()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Completed populating database (%s)\n", getTimeString())
}

func getTimeString() string {
	currentTime := time.Now()
	return fmt.Sprintf("%d:%02d:%d.%d", currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond())
}
