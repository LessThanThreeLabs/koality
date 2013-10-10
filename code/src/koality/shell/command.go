package shell

import (
	"fmt"
	"strings"
)

type Command string

func Group(command Command) Command {
	return Command(fmt.Sprintf("{\n%s\n}", command))
}

func Advertised(command Command) Command {
	return AdvertisedWithActual(command, command)
}

func AdvertisedWithActual(advertised, actual Command) Command {
	colorize := func(line string) string {
		return fmt.Sprintf("\\x1b[34m%s\\x1b[0m", line)
	}
	lines := strings.Split(string(advertised), "\n")
	colorized := make([]string, len(lines))
	for index, line := range lines {
		colorized[index] = colorize(line)
	}
	printCommand := Command(fmt.Sprintf("printf %s", Quote(fmt.Sprintf("$ %s\\n", strings.Join(colorized, "\\n> ")))))
	return And(printCommand, actual)
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

func Not(command Command) Command {
	return Command(fmt.Sprintf("! %s", command))
}

func And(commands ...Command) Command {
	commandStrings := make([]string, len(commands))
	for index, value := range commands {
		commandStrings[index] = string(value)
	}
	return Command(strings.Join(commandStrings, " && "))
}

func Or(commands ...Command) Command {
	commandStrings := make([]string, len(commands))
	for index, value := range commands {
		commandStrings[index] = string(value)
	}
	return Command(strings.Join(commandStrings, " || "))
}

func Pipe(commands ...Command) Command {
	commandStrings := make([]string, len(commands))
	for index, value := range commands {
		commandStrings[index] = string(value)
	}
	return Command(strings.Join(commandStrings, " | "))
}

func Chain(commands ...Command) Command {
	commandStrings := make([]string, len(commands))
	for index, value := range commands {
		commandStrings[index] = string(value)
	}
	return Command(strings.Join(commandStrings, "\n"))
}

func Background(command Command) Command {
	return Command(fmt.Sprintf("%s &", command))
}

func Subshell(command Command) Command {
	return Command(fmt.Sprintf("(%s)", command))
}

func Capture(command Command) Command {
	return Command(fmt.Sprintf("$(%s)", command))
}

func Redirect(command, outfile Command, includeStderr bool) Command {
	redirect := fmt.Sprintf("%s > %s", Group(command), outfile)
	if includeStderr {
		return Command(fmt.Sprintf("%s 2>&1", redirect))
	} else {
		return Command(redirect)
	}
}

func Append(command, outfile Command, includeStderr bool) Command {
	redirect := fmt.Sprintf("%s >> %s", Group(command), outfile)
	if includeStderr {
		return Command(fmt.Sprintf("%s 2>&1", redirect))
	} else {
		return Command(redirect)
	}
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
