package remotecommand

import (
	"fmt"
	"koality/shell"
)

type RemoteCommand struct {
	advertised bool
	directory  string
	name       string
	timeout    int
	xunit      []string
	commands   []string
}

func NewRemoteCommand(advertised bool, directory, name string, timeout int, xunit, commands []string) (remoteCommand RemoteCommand) {
	remoteCommand.advertised = advertised
	remoteCommand.directory = directory
	remoteCommand.name = name
	remoteCommand.timeout = timeout
	remoteCommand.xunit = xunit
	remoteCommand.commands = commands
	return
}

func (remoteCommand RemoteCommand) Name() string {
	return remoteCommand.name
}

func (remoteCommand RemoteCommand) Executable() shell.Executable {
	return shell.Executable{
		Command: remoteCommand.toScript(),
	}
}

func (remoteCommand RemoteCommand) XunitPaths() []string {
	return remoteCommand.xunit
}

func advertiseCommand(remoteCommand *RemoteCommand, command string) shell.Command {
	if remoteCommand.advertised {
		return shell.Advertised(shell.Command(command))
	} else {
		return shell.Command(command)
	}
}

func advertiseCommands(remoteCommand *RemoteCommand) shell.Command {
	var commands []shell.Command

	for _, command := range remoteCommand.commands {
		commands = append(commands, advertiseCommand(remoteCommand, command))
	}
	return shell.And(commands...)
}

func (remoteCommand *RemoteCommand) toScript() shell.Command {
	givenCommand := shell.Command(fmt.Sprintf("eval %s", shell.Quote(string(shell.And(advertiseCommands(remoteCommand))))))

	timeoutMessage := fmt.Sprintf("echo %s timed out after %d seconds", shell.Quote(remoteCommand.name), remoteCommand.timeout)

	timeoutCommand := shell.Chain(
		shell.Commandf("sleep %d", remoteCommand.timeout),
		shell.Silent("kill -INT $$"),
		shell.Command("sleep 1"),
		shell.IfElse(
			shell.Silent("kill -0 $$"),
			shell.Chain(
				shell.Command("sleep 2"),
				shell.Silent("kill -KILL $$"),
				shell.Command("echo"),
				shell.Command(timeoutMessage),
				shell.Command("kill -9 0"),
			),
			shell.Chain(
				shell.Commandf("echo"),
				shell.Command(timeoutMessage),
				shell.Command("kill -9 0"),
			),
		),
	)

	commandsWithTimeout := shell.Chain(
		shell.Background(shell.Silent(shell.Subshell(timeoutCommand))),
		shell.Command("watchdogpid=$!"),
		givenCommand,
		shell.Command("_r=$?"),
		shell.Command("exec 2>/dev/null"),
		shell.Silent("kill -KILL $watchdogpid"),
		shell.Silent("pkill -KILL -P $watchdogpid"),
		shell.Or(
			shell.Test("$_r -eq 0"),
			shell.Command(fmt.Sprintf("echo %s failed with return code $_r", shell.Quote(remoteCommand.name))),
		),
		shell.Command("exit $_r"),
	)

	return shell.Login(shell.And(
		advertiseCommand(remoteCommand, fmt.Sprintf("cd %s", remoteCommand.directory)),
		commandsWithTimeout,
	))
}
