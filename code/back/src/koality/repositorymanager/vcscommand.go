package repositorymanager

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
	command := exec.Command(repository.getVcsBaseCommand(), arguments...)

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	command.Stdout, command.Stderr = stdout, stderr

	command.Dir = repository.getPath()
	command.Env = env

	return &VcsCommand{command, stdout, stderr}
}

func RunCommand(vcsCommand *VcsCommand) error {
	err := vcsCommand.Command.Run()
	if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("Attempting to run command %v resulted in a non-zero return state.", vcsCommand.Command)
	} else if err != nil {
		return err
	}

	return nil
}
