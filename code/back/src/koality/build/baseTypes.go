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

type basicCommand struct {
	name       string
	executable shell.Executable
}

func NewBasicCommand(name string, executable shell.Executable) Command {
	return &basicCommand{name, executable}
}

func (command *basicCommand) Name() string {
	return command.name
}

func (command *basicCommand) Executable() shell.Executable {
	return command.executable
}

func (command *basicCommand) XunitPaths() []string {
	return nil
}

// Temporary

func NewShellCommand(name string, command shell.Command) Command {
	return NewBasicCommand(name, shell.Executable{
		Command: command,
	})
}
