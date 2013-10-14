package vm

import (
	"fmt"
	"koality/shell"
	"os/exec"
	"strconv"
	"strings"
)

type SshCaller interface {
	SshCall(command shell.Command) exec.Cmd
}

type SshExecutor struct {
	SshConfig
	Executor
}

func (sshExecutor SshExecutor) Execute(command shell.Command) exec.Cmd {
	fullCommand := shell.Command(strings.Join(append(sshExecutor.SshConfig.ArgList(), string(command)), " "))
	return sshExecutor.Executor.Execute(fullCommand)
}

type ScpCaller interface {
	ScpCall(localFilePath, remoteFilePath string, retrieveFile bool) exec.Cmd
}

type SshConfig struct {
	Username string
	Hostname string
	Port     int
	Options  map[string]string
}

type ScpConfig struct {
	SshConfig
	LocalFilePath  string
	RemoteFilePath string
	RetrieveFile   bool
}

func toOptionsList(options map[string]string) []string {
	optionsList := make([]string, len(options))

	index := 0
	for key, value := range options {
		optionsList[index] = fmt.Sprintf("-o%s=%s", key, value)
		index++
	}
	return optionsList
}

func (sshConfig SshConfig) ArgList() []string {
	options := toOptionsList(sshConfig.Options)
	login := fmt.Sprintf("%s@%s", sshConfig.Username, sshConfig.Hostname)
	args := append(options, login, "-p", strconv.Itoa(sshConfig.Port))

	return append([]string{"ssh"}, args...)
}

func (scpConfig ScpConfig) ArgList() []string {
	options := toOptionsList(scpConfig.SshConfig.Options)
	remotePath := fmt.Sprintf("%s@%s:%s", scpConfig.SshConfig.Username, scpConfig.SshConfig.Hostname, scpConfig.RemoteFilePath)

	if scpConfig.RetrieveFile {
		return append(options, "-P", strconv.Itoa(scpConfig.SshConfig.Port), remotePath, scpConfig.LocalFilePath)
	} else {
		return append(options, "-P", strconv.Itoa(scpConfig.SshConfig.Port), scpConfig.LocalFilePath, remotePath)
	}
}
