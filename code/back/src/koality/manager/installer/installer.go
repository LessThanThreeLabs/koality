package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"koality/shell"
	"koality/util/pathtranslator"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var requiredLibraries = []string{"curl"}

type koalityMetadata struct {
	Version string `json:"version"`
	Release string `json:"release"`
}

func main() {
	if err := installKoality(); err != nil {
		panic(err)
	}
	fmt.Println("Koality installed successfully")
}

func installKoality() error {
	baseDirectory, err := pathtranslator.TranslatePath(".")
	if err != nil {
		return err
	}

	metadataBytes, err := ioutil.ReadFile(filepath.Join(baseDirectory, ".metadata"))
	if err != nil {
		return err
	}

	var metadata koalityMetadata
	if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
		return err
	}

	if metadata.Version == "" {
		return fmt.Errorf("Version not specified in metadata file")
	}

	version := metadata.Version
	if metadata.Release == "development" {
		version = fmt.Sprintf("%s-dev-%d", metadata.Version, time.Now().Unix())
	}

	installDirectory := filepath.Join("/", "etc", "koality", "install", version)

	if _, err = os.Stat(installDirectory); err == nil {
		if err = os.RemoveAll(installDirectory); err != nil {
			return fmt.Errorf("Failed to delete the previous install directory at %s\nError: %v", installDirectory, err)
		}
	}

	if err = os.MkdirAll(filepath.Dir(installDirectory), 0755); err != nil {
		return fmt.Errorf("Failed to create the install directory at %s\nError: %v", installDirectory, err)
	}

	moveCommand := exec.Command("mv", baseDirectory, installDirectory)
	moveStdout, moveStderr := new(bytes.Buffer), new(bytes.Buffer)
	moveCommand.Stdout, moveCommand.Stderr = moveStdout, moveStderr
	if err := moveCommand.Run(); err != nil {
		return fmt.Errorf("Failed to move Koality to the install directory at %s\nError: %v", installDirectory, err)
	}

	if _, err := user.Lookup("nginx"); err != nil {
		addNginxUserCommand := exec.Command("bash", "-c", string(shell.Sudo(
			shell.Commandf("useradd nginx"),
		)))
		addNginxUserStdout, addNginxUserStderr := new(bytes.Buffer), new(bytes.Buffer)
		addNginxUserCommand.Stdout, addNginxUserCommand.Stderr = addNginxUserStdout, addNginxUserStderr
		if err = addNginxUserCommand.Run(); err != nil {
			return fmt.Errorf("Failed to add nginx user\nStdout: %s\nStderr: %s\nError: %v", addNginxUserStdout.String(), addNginxUserStderr.String(), err)
		}
	}

	installDependenciesCommand := exec.Command("bash", "-c", string(shell.Sudo(
		shell.Commandf("apt-get install -y %s", strings.Join(requiredLibraries, " ")),
	)))
	installDependenciesStdout, installDependenciesStderr := new(bytes.Buffer), new(bytes.Buffer)
	installDependenciesCommand.Stdout, installDependenciesCommand.Stderr = installDependenciesStdout, installDependenciesStderr
	if err := installDependenciesCommand.Run(); err != nil {
		return fmt.Errorf("Failed to install dependencies: %v\nStdout: %s\nStderr: %s\nError: %v", requiredLibraries, installDependenciesStdout.String(), installDependenciesStderr.String(), err)
	}

	postgresqlSourceFile := filepath.Join("/", "etc", "apt", "sources.list.d", "pgdg.list")
	if _, err := os.Stat(postgresqlSourceFile); err != nil {
		if err = ioutil.WriteFile(postgresqlSourceFile, []byte("deb http://apt.postgresql.org/pub/repos/apt/ precise-pgdg main"), 0644); err != nil {
			return fmt.Errorf("Failed to create the postgresql sources file at %s\nError: %v", postgresqlSourceFile, err)
		}
	}

	addPostgresqlKeysCommand := exec.Command("bash", "-c", string(shell.Pipe(
		shell.Command("curl https://www.postgresql.org/media/keys/ACCC4CF8.asc"),
		shell.Sudo("apt-key add -"),
	)))
	addPostgresqlKeysStdout, addPostgresqlKeysStderr := new(bytes.Buffer), new(bytes.Buffer)
	addPostgresqlKeysCommand.Stdout, addPostgresqlKeysCommand.Stderr = addPostgresqlKeysStdout, addPostgresqlKeysStderr
	if err := addPostgresqlKeysCommand.Run(); err != nil {
		return fmt.Errorf("Failed to install postgresql keys\nStdout: %s\nStderr: %s\nError: %v", addPostgresqlKeysStdout.String(), addPostgresqlKeysStderr.String(), err)
	}

	aptUpdateCommand := exec.Command("bash", "-c", string(shell.Sudo("apt-get update")))
	aptUpdateStdout, aptUpdateStderr := new(bytes.Buffer), new(bytes.Buffer)
	aptUpdateCommand.Stdout, aptUpdateCommand.Stderr = aptUpdateStdout, aptUpdateStderr
	if err := aptUpdateCommand.Run(); err != nil {
		return fmt.Errorf("Failed to apt-get update\nStdout: %s\nStderr: %s\nError: %v", aptUpdateStdout.String(), aptUpdateStderr.String(), err)
	}

	installPostgresqlCommand := exec.Command("bash", "-c", string(shell.Sudo("apt-get install -y postgresql-9.3")))
	installPostgresqlStdout, installPostgresqlStderr := new(bytes.Buffer), new(bytes.Buffer)
	installPostgresqlCommand.Stdout, installPostgresqlCommand.Stderr = installPostgresqlStdout, installPostgresqlStderr
	if err := installPostgresqlCommand.Run(); err != nil {
		return fmt.Errorf("Failed to install postgresql\nStdout: %s\nStderr: %s\nError: %v", installPostgresqlStdout.String(), installPostgresqlStderr.String(), err)
	}

	configurePostgresqlCommand := exec.Command("bash", "-c", string(
		shell.And(
			shell.Or(
				shell.Pipe(shell.AsUser("postgres", shell.Commandf("psql -c %s", shell.Quote("SELECT 1 FROM pg_user WHERE usename='koality'"))), shell.Command("grep -q 1")),
				shell.AsUser("postgres", shell.Commandf("psql -c %s", shell.Quote("CREATE USER koality PASSWORD 'lt3' SUPERUSER"))),
			),
			shell.Or(
				shell.Pipe(shell.AsUser("postgres", shell.Commandf("psql -c %s", shell.Quote("SELECT 1 FROM pg_database WHERE datname='koality'"))), shell.Command("grep -q 1")),
				shell.AsUser("postgres", shell.Command("createdb koality --template template0 --locale en_US.utf8 --encoding UTF8")),
			),
		),
	))
	configurePostgresqlStdout, configurePostgresqlStderr := new(bytes.Buffer), new(bytes.Buffer)
	configurePostgresqlCommand.Stdout, configurePostgresqlCommand.Stderr = configurePostgresqlStdout, configurePostgresqlStderr
	if err := configurePostgresqlCommand.Run(); err != nil {
		return fmt.Errorf("Failed to configure postgresql\nStdout: %s\nStderr: %s\nError: %v", configurePostgresqlStdout.String(), configurePostgresqlStderr.String(), err)
	}

	installOpenSshCommand := exec.Command("bash", "-c", string(shell.Sudo("make install")))
	installOpenSshCommand.Dir = filepath.Join(installDirectory, "dependencies", "openssh-6.0p1")
	installOpenSshStdout, installOpenSshStderr := new(bytes.Buffer), new(bytes.Buffer)
	installOpenSshCommand.Stdout, installOpenSshCommand.Stderr = installOpenSshStdout, installOpenSshStderr
	if err := installOpenSshCommand.Run(); err != nil {
		return fmt.Errorf("Failed to install OpenSSH\nStdout: %s\nStderr: %s\nError: %v", installOpenSshStdout.String(), installOpenSshStderr.String(), err)
	}

	relinkCommand := exec.Command("bash", "-c", string(shell.Sudo(shell.Commandf("ln -n -f -s %s %s", installDirectory, filepath.Join("/", "etc", "koality", "current")))))
	relinkStdout, relinkStderr := new(bytes.Buffer), new(bytes.Buffer)
	relinkCommand.Stdout, relinkCommand.Stderr = relinkStdout, relinkStderr
	if err := relinkCommand.Run(); err != nil {
		return fmt.Errorf("Failed to link new Koality version at %s\nStdout: %s\nStderr: %s\nError: %v", installDirectory, relinkStdout.String(), relinkStderr.String(), err)
	}

	return nil
}