package main

import (
	"fmt"
	"io/ioutil"
	"koality/util/pathtranslator"
	"koality/util/ssh"
	"log"
	"os"
)

func main() {
	publicKeyBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	shellPath, err := pathtranslator.TranslatePathAndCheckExists(pathtranslator.BinaryPath("restrictedShell"))
	if err != nil {
		log.Fatal(err)
	}

	cmd, err := ssh.GetForcedCommand(shellPath, string(publicKeyBytes))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(cmd)
}
