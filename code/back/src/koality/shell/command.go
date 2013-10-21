package shell

import (
	"fmt"
	"strings"
)

type Command string

func Group(command Command) Command {
	return Command(fmt.Sprintf("{\n%s\n}", command))
}

func Subshell(command Command) Command {
	return Command(fmt.Sprintf("(%s)", command))
}

func Capture(command Command) Command {
	return Command(fmt.Sprintf("$(%s)", command))
}

func Background(command Command) Command {
	return Command(fmt.Sprintf("%s &", command))
}

func Not(command Command) Command {
	return Command(fmt.Sprintf("! %s", Group(command)))
}

func Test(command Command) Command {
	return Command(fmt.Sprintf("[ %s ]", command))
}

func If(condition, thenCommand Command) Command {
	return Chain(
		Command(fmt.Sprintf("if %s", condition)),
		Command("then"),
		thenCommand,
		Command("fi"),
	)
}

func IfElse(condition, thenCommand, elseCommand Command) Command {
	return Chain(
		Command(fmt.Sprintf("if %s", condition)),
		Command("then"),
		thenCommand,
		Command("else"),
		elseCommand,
		Command("fi"),
	)
}

func join(commands []Command, joiner string) Command {
	commandStrings := make([]string, len(commands))
	for index, command := range commands {
		commandStrings[index] = string(Group(command))
	}
	return Command(strings.Join(commandStrings, joiner))
}

func And(commands ...Command) Command {
	return join(commands, " && ")
}

func Or(commands ...Command) Command {
	return join(commands, " || ")
}

func Pipe(commands ...Command) Command {
	return join(commands, " | ")
}

func Chain(commands ...Command) Command {
	return join(commands, "\n")
}

func redirect(command, stdoutFile Command, includeStderr bool, redirectionType string) Command {
	redirect := fmt.Sprintf("%s %s %s", Group(command), redirectionType, stdoutFile)
	if includeStderr {
		return Command(fmt.Sprintf("%s 2>&1", redirect))
	} else {
		return Command(redirect)
	}
}

func Redirect(command, stdoutFile Command, includeStderr bool) Command {
	return redirect(command, stdoutFile, includeStderr, ">")
}

func Append(command, stdoutFile Command, includeStderr bool) Command {
	return redirect(command, stdoutFile, includeStderr, ">>")
}

func Silent(command Command) Command {
	return Redirect(command, Command("/dev/null"), true)
}

func Login(command Command) Command {
	return Command(fmt.Sprintf("bash --login -c %s", Quote(string(command))))
}

func Sudo(command Command) Command {
	return Command(fmt.Sprintf("sudo -E HOME=\"$HOME\" PATH=\"$PATH\" bash -c %s", Quote(string(command))))
}

func Advertised(command Command) Command {
	return AdvertisedWithActual(command, command)
}

func AdvertisedWithActual(advertised, actual Command) Command {
	colorize := func(line string) string {
		return fmt.Sprintf("%s%s%s", AnsiFormat(AnsiFgBlue), line, AnsiFormat(AnsiReset))
	}
	lines := strings.Split(string(advertised), "\n")
	colorized := make([]string, len(lines))
	for index, line := range lines {
		colorized[index] = colorize(line)
	}
	printCommand := Command(fmt.Sprintf("printf %s", Quote(fmt.Sprintf("$ %s\\n", strings.Join(colorized, "\\n> ")))))
	return And(printCommand, actual)
}