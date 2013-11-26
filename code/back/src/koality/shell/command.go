package shell

import (
	"fmt"
	"strings"
)

type Command string

func Commandf(format string, args ...interface{}) Command {
	return Command(fmt.Sprintf(format, args...))
}

func Group(command Command) Command {
	return Commandf("{\n%s\n}", command)
}

func Subshell(command Command) Command {
	return Commandf("(%s)", command)
}

func Capture(command Command) Command {
	return Commandf("$(%s)", command)
}

func Background(command Command) Command {
	return Commandf("%s &", command)
}

func Not(command Command) Command {
	return Commandf("! %s", Group(command))
}

func Test(command Command) Command {
	return Commandf("[ %s ]", command)
}

func If(condition, thenCommand Command) Command {
	return Commandf("if %s; then\n%s\nfi", condition, thenCommand)
}

func IfElse(condition, thenCommand, elseCommand Command) Command {
	return Commandf("if %s; then\n%s\nelse\n%s\nfi", condition, thenCommand, elseCommand)
}

func join(commands []Command, joiner string, grouped bool) Command {
	commandStrings := make([]string, len(commands))
	for index, command := range commands {
		if grouped {
			commandStrings[index] = string(Group(command))
		} else {
			commandStrings[index] = string(command)
		}
	}
	return Command(strings.Join(commandStrings, joiner))
}

func And(commands ...Command) Command {
	return join(commands, " && ", true)
}

func Or(commands ...Command) Command {
	return join(commands, " || ", true)
}

func Pipe(commands ...Command) Command {
	return join(commands, " | ", true)
}

func Chain(commands ...Command) Command {
	return join(commands, "\n", false)
}

func redirect(command, stdoutFile Command, includeStderr bool, redirectionType string) Command {
	redirect := fmt.Sprintf("%s %s %s", Group(command), redirectionType, stdoutFile)
	if includeStderr {
		return Commandf("%s 2>&1", redirect)
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
	return Commandf("bash --login -c %s", Quote(string(command)))
}

func Sudo(command Command) Command {
	return Commandf("sudo -E HOME=\"$HOME\" PATH=\"$PATH\" bash -c %s", Quote(string(command)))
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
	printCommand := Commandf("printf %s", Quote(fmt.Sprintf("$ %s\\n", strings.Join(colorized, "\\n> "))))
	return And(printCommand, actual)
}
