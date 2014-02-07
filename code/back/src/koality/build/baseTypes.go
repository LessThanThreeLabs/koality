package build

import (
	"koality/shell"
)

type Command interface {
	Name() string
	Executable() shell.Executable
	XunitPaths() []string
}

type SectionResult struct {
	Section       string
	Final         bool
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

func (shellCommand ShellCommand) Executable() shell.Executable {
	return shell.Executable{
		Command: shellCommand.command,
	}
}

func (shellCommand ShellCommand) XunitPaths() []string {
	return nil
}
