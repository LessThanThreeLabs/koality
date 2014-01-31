package test

import (
	"koality/internalapi"
	"koality/resources"
	"koality/vm"
	"time"
)

const (
	PrivateKey       = "pkey"
	RepositoriesPath = "/etc/koality/repositories"
)

var (
	currentTime = time.Now()
	Repository  = resources.Repository{
		Id:        100,
		Name:      "koality-v1",
		RemoteUri: "git@github.com:KoalityCode/koality-v1.git",
		Created:   &currentTime,
	}
)

type PublicKeyVerifier struct{}

func (mockPublicKeyVerifier PublicKeyVerifier) GetUserIdForKey(publicKey string, userId *uint64) (err error) {
	*userId = 69
	return
}

type RepositoryReader struct{}

func (repositoryReader RepositoryReader) GetRepoByName(name string, repositoryInfo *internalapi.RepositoryInfo) error {
	repositoryInfo.Repository = Repository
	repositoryInfo.RepositoriesPath = RepositoriesPath
	return nil
}

type UserInfoReader struct{}

func (userInfoReader UserInfoReader) GetRepoPrivateKey(_ interface{}, privateKeyRes *string) error {
	*privateKeyRes = PrivateKey
	return nil
}

type VmReader struct{}

func (vmReader VmReader) GetShellCommandFromId(args internalapi.VmReaderArg, commandRes *vm.Command) error {
	username := "koality"
	hostname := "koalitycode.com"
	sshConfig := vm.SshConfig{
		Username:   username,
		Hostname:   hostname,
		Port:       22,
		PrivateKey: PrivateKey,
		Options: map[string]string{
			"LogLevel":              "error",
			"StrictHostKeyChecking": "no",
			"UserKnownHostsFile":    "/dev/null",
			"ServerAliveInterval":   "20",
		},
	}
	*commandRes = vm.Command{
		Argv: sshConfig.SshArgs(""),
		Envv: []string{"PRIVATE_KEY=" + sshConfig.PrivateKey},
	}
	return nil
}
