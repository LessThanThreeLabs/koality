package main

import (
	"fmt"
	"io/ioutil"
	"koality/util/ssh"
	"log"
	"os"
	"path"
	"path/filepath"
)

func main() {
	publicKeyBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	absPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}

	serve := path.Join(filepath.Dir(absPath), "restrictedShell")
	cmd, err := ssh.GetForcedCommand(serve, string(publicKeyBytes))
	fmt.Printf(cmd)
}
