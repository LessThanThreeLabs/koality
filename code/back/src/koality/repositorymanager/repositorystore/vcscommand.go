package repositorystore

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Repository interface {
	getVcsBaseCommand() string
	getPath() string
}

type VcsCommand struct {
	Command *exec.Cmd
	Stdout  *bytes.Buffer
	Stderr  *bytes.Buffer
}

func Command(repository Repository, Env []string, cmd string, args ...string) *VcsCommand {
	arguments := append([]string{cmd}, args...)
	arguments = append(arguments, args...)
	command := exec.Command(repository.getVcsBaseCommand(), arguments...)

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	command.Stdout, command.Stderr = stdout, stderr

	command.Dir = repository.getPath()

	return &VcsCommand{command, stdout, stderr}
}

func RunCommand(vcsCommand *VcsCommand) (err error) {
	err = vcsCommand.Command.Run()
	if success := vcsCommand.Command.ProcessState.Success(); !success {
		return fmt.Errorf("Attempting to run command %v resulted in a non-zero return state.", vcsCommand.Command)
	}
	return
}
