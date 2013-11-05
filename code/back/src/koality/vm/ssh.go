package vm

import (
	"fmt"
	"koality/shell"
	"strconv"
	"strings"
)

type SshExecutableMaker struct {
	sshConfig       SshConfig
	executableMaker shell.ExecutableMaker
}

func NewSshExecutableMaker(config SshConfig) *SshExecutableMaker {
	return &SshExecutableMaker{
		sshConfig:       config,
		executableMaker: shell.NewShellExecutableMaker(),
	}
}

func (sshExecutableMaker *SshExecutableMaker) MakeExecutable(command shell.Command) shell.Executable {
	fullCommand := shell.Command(strings.Join(append(sshExecutableMaker.sshConfig.SshArgs(string(command))), " "))
	return sshExecutableMaker.executableMaker.MakeExecutable(fullCommand)
}

type Scper interface {
	Scp(localFilePath, remoteFilePath string, retrieveFile bool) shell.Executable
}

type ShellScper struct {
	scpConfig       ScpConfig
	executableMaker shell.ExecutableMaker
}

func NewScper(config ScpConfig) Scper {
	return &ShellScper{
		scpConfig:       config,
		executableMaker: shell.NewShellExecutableMaker(),
	}
}

func (shellScper *ShellScper) Scp(localFilePath, remoteFilePath string, retrieveFile bool) shell.Executable {
	fullCommand := shell.Command(strings.Join(append(shellScper.scpConfig.ScpArgs(localFilePath, remoteFilePath, retrieveFile)), " "))
	return shellScper.executableMaker.MakeExecutable(fullCommand)
}

type SshConfig struct {
	Username string
	Hostname string
	Port     int
	Options  map[string]string
}

type ScpConfig SshConfig

func toOptionsList(options map[string]string) []string {
	optionsList := make([]string, len(options))

	index := 0
	for key, value := range options {
		optionsList[index] = fmt.Sprintf("-o%s=%s", key, value)
		index++
	}
	return optionsList
}

func (sshConfig SshConfig) SshArgs(remoteCommand string) []string {
	options := toOptionsList(sshConfig.Options)
	login := fmt.Sprintf("%s@%s", sshConfig.Username, sshConfig.Hostname)
	args := append(options, login, "-p", strconv.Itoa(sshConfig.Port), shell.Quote(remoteCommand))

	return append([]string{"ssh"}, args...)
}

func (scpConfig ScpConfig) ScpArgs(localFilePath, remoteFilePath string, retrieveFile bool) []string {
	options := toOptionsList(scpConfig.Options)
	remotePath := fmt.Sprintf("%s@%s:%s", scpConfig.Username, scpConfig.Hostname, remoteFilePath)

	if retrieveFile {
		return append(append([]string{"scp"}, options...), "-P", strconv.Itoa(scpConfig.Port), shell.Quote(remotePath), shell.Quote(localFilePath))
	} else {
		return append(append([]string{"scp"}, options...), "-P", strconv.Itoa(scpConfig.Port), shell.Quote(localFilePath), shell.Quote(remotePath))
	}
}

type ScpFileCopier struct {
	Scper
}

func (fileCopier *ScpFileCopier) FileCopy(localFilePath, remoteFilePath string) (shell.Executable, error) {
	return fileCopier.Scper.Scp(localFilePath, remoteFilePath, false), nil
}
