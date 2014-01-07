package main

import (
	"fmt"
	"koality/resources/database"
	"koality/webserver"
)

const (
	webserverPort = 8080
)

func main() {
	resourcesConnection, err := database.New()
	if err != nil {
		panic(err)
	}

	// TODO: initialize more components here

	fmt.Println("Koality successfully started!")

	// This will block
	err = webserver.Start(resourcesConnection, webserverPort)
	if err != nil {
		panic(err)
	}
}
