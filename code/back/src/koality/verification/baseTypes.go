package verification

import (
	"koality/shell"
)

type Command interface {
	Name() string
	ShellCommand() shell.Command
}

type TestCommand interface {
	Command
	GetXunitCommand() shell.Command
}

type Result struct {
	StageType string
	Passed    bool
}

type ChangeStatus struct {
	Failed    bool
	Cancelled bool
}

// Temporary

type ShellCommand struct {
	name    string
	command shell.Command
}

func NewShellCommand(name string, command shell.Command) (shellCommand ShellCommand) {
	shellCommand.name = name
	shellCommand.command = command
	return
}

func (shellCommand ShellCommand) Name() string {
	return shellCommand.name
}

func (shellCommand ShellCommand) ShellCommand() shell.Command {
	return shellCommand.command
}
