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

type SectionResult struct {
	Section       string
	FailSectionOn string
	Passed        bool
}

// Temporary

type ShellCommand struct {
	name    string
	command shell.Command
}

func NewShellCommand(name string, command shell.Command) ShellCommand {
	return ShellCommand{name, command}
}

func (shellCommand ShellCommand) Name() string {
	return shellCommand.name
}

func (shellCommand ShellCommand) ShellCommand() shell.Command {
	return shellCommand.command
}
