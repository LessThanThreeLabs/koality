package main

import (
	"koality/util/ssh"
	"log"
	"os"
	"strconv"
)

func main() {
	userId, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(ssh.HandleSSHCommand(userId, os.Getenv("SSH_ORIGINAL_COMMAND")))
}
