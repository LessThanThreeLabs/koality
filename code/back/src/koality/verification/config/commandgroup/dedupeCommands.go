package commandgroup

import (
	"fmt"
	"koality/shell"
	"koality/verification"
)

func dedupeCommands(commands []verification.Command) []verification.Command {
	newCommands := make([]verification.Command, len(commands))
	nameCounts := make(map[string]int, len(commands))

	for index, command := range commands {
		deduped, ok := command.(dedupedCommand)
		if ok {
			command = deduped.command
		}

		nameCount := nameCounts[command.Name()]
		nameCounts[command.Name()] = nameCount + 1

		if nameCount > 0 {
			command = dedupedCommand{command, nameCount + 1}
		}
		newCommands[index] = command
	}
	return newCommands
}

type dedupedCommand struct {
	command      verification.Command
	suffixNumber int
}

func (command dedupedCommand) Name() string {
	return fmt.Sprintf("%s #%d", command.command.Name(), command.suffixNumber)
}

func (command dedupedCommand) ShellCommand() shell.Command {
	return command.command.ShellCommand()
}
