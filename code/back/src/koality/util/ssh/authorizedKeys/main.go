package main

import (
	"fmt"
	"io/ioutil"
	"koality/util/ssh"
	"log"
	"os"
	"os/exec"
)

func main() {
	publicKeyBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	shellPath, err := exec.LookPath("restrictedShell")
	if err != nil {
		log.Fatal(err)
	}

	cmd, err := ssh.GetForcedCommand(shellPath, string(publicKeyBytes))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(cmd)
}
