package repositorystore

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type Repository struct {
	vcsBaseCommand string
	path           string
}

type VcsCommand struct {
	Command *exec.Cmd
	Stdout  *bytes.Buffer
	Stderr  *bytes.Buffer
}

func Open(vcsType string, path string) (vcsRepository *Repository, err error) {
	// TODO(akostov) the keys for this map should be constants defined in resources.
	vcsDispatcher := map[string]string{
		"git": "git",
		"hg":  "hg",
	}

	vcsRepository.vcsBaseCommand = vcsDispatcher[vcsType]

	if _, err = os.Stat(path); os.IsNotExist(err) {
		return nil, NoSuchRepositoryInStoreError{fmt.Sprintf("The path %v does not exist.", path)}
	}
	vcsRepository.path = path

	return
}

func (repository *Repository) Command(Env []string, cmd string, args ...string) *VcsCommand {
	var arguments []string

	arguments = append(arguments, cmd)
	arguments = append(arguments, args...)
	command := exec.Command(repository.vcsBaseCommand, arguments...)

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	command.Stdout, command.Stderr = stdout, stderr

	command.Dir = repository.path

	return &VcsCommand{command, stdout, stderr}
}

func RunCommand(command *VcsCommand) (success bool, err error) {
	if err = command.Command.Run(); err != nil {
		return
	}

	success = command.Command.ProcessState.Success()
	return
}
