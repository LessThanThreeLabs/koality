package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"koality/license/server"
	"koality/util/pathtranslator"
	"path"
)

const (
	port = 9000
)

func main() {
	// TODO (bbland): figure out what path to use here
	configPath, err := pathtranslator.TranslatePathAndCheckExists(path.Join(
		"code", "back", "src", "koality", "license", "server", "licenseServer", "license-server-config.json"))
	if err != nil {
		panic(err)
	}
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	config := new(licenseserver.DatabaseConfiguration)
	if err = json.Unmarshal(configBytes, config); err != nil {
		panic(err)
	}

	database, err := licenseserver.GetDatabaseConnection(config)
	if err != nil {
		panic(err)
	}

	licenseServer := licenseserver.New(database, port)

	fmt.Printf("Starting license server on port %d...\n", port)

	if err = licenseServer.Start(); err != nil {
		panic(err)
	}
}
