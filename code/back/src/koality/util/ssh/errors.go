package ssh

import (
	"fmt"
)

type InvalidCommandError struct {
	command string
}

func (err InvalidCommandError) Error() string {
	return fmt.Sprintf(`"%s" cannot be executed in this restricted shell.`, err.command)
}

type MalformedCommandError struct {
	repoPath string
}

func (err MalformedCommandError) Error() string {
	return fmt.Sprintf(`repository path: %s. Repository path cannot contain "..".`, err.repoPath)
}

type RepositoryNotFoundError struct {
}

func (err RepositoryNotFoundError) Error() string {
	return "Repository not found. Please check your VCS' configuration."
}

type UserNotFoundError struct {
	userId uint64
}

func (err UserNotFoundError) Error() string {
	return fmt.Sprintf("No user exists with id %d.", err.userId)
}

type NoShellAccessError struct {
	userEmail string
}

func (err NoShellAccessError) Error() string {
	return fmt.Sprintf("You have been successfully authenticated as %s, but shell access is not permitted", err.userEmail)
}

type InvalidPermissionsError struct {
	userId  uint64
	command string
	repo_id uint64
}

func (err InvalidPermissionsError) Error() string {
	if err.repo_id == 0 {
		return fmt.Sprintf("User %d does not have the necessary permissions to run %s on repository",
			err.userId, err.command)
	} else {
		return fmt.Sprintf("User %d does not have the necessary permissions to run %s on repository %d",
			err.userId, err.command, err.repo_id)
	}
}

type VirtualMachineNotFoundError struct {
	instanceId string
}

func (err VirtualMachineNotFoundError) Error() string {
	return fmt.Sprintf(`Could not find virtual machine with id "%s"`, err.instanceId)
}
