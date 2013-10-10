package main

import (
	"fmt"
	"koality/resources/amqp"
	"log"
)

func main() {
	fmt.Printf("Hello, world.\n")

	connection := amqp.NewConnection()
	fmt.Printf("%v\n\n", connection)

	user, error := connection.Users.Read.Get(17)
	if error != nil {
		log.Fatalf("%s", error)
	}

	fmt.Printf("%v\n\n", user)
}
