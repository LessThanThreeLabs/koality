package verification

import (
	"koality/shell"
	"strings"
)

type Command interface {
	Name() string
	ShellCommand() shell.Command
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
	shell.Command
}

func (shellCommand ShellCommand) Name() string {
	return strings.Fields(string(shellCommand.Command))[0]
}

func (shellCommand ShellCommand) ShellCommand() shell.Command {
	return shellCommand.Command
}
