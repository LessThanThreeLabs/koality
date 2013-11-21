package verification

import (
	"koality/shell"
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
