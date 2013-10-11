package main

import (
	"fmt"
	"koality/resources/rpc/client"
)

func main() {
	connection := client.NewConnection()

	fmt.Printf("Getting user 17...\n")
	user, err := connection.Users.Read.Get(17)
	if err != nil {
		fmt.Println("Error: %v", err)
	} else {
		fmt.Println(user)
	}
}
