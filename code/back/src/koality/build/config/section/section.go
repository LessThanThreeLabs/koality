package section

import (
	"fmt"
	"koality/build"
	"koality/build/config/commandgroup"
)

// mandatory
const (
	RunOnAll    = "runOnAll"
	RunOnSplit  = "runOnSplit"
	RunOnSingle = "runOnSingle"
)

const (
	FailOnNever = "failOnNever"
	FailOnAny   = "failOnAny"
	FailOnFirst = "failOnFirst"
)

type section struct {
	name              string
	isFinal           bool
	runOn             string
	failOn            string
	continueOnFailure bool
	factoryCommands   commandgroup.CommandGroup
	commands          commandgroup.AppendableCommandGroup
	exportPaths       []string
}

type Section interface {
	Name() string
	IsFinal() bool
	FailOn() string
	ContinueOnFailure() bool
	FactoryCommands(readOnlyCopy bool) commandgroup.CommandGroup
	Commands(readOnlyCopy bool) commandgroup.CommandGroup
	AppendCommand(build.Command) (build.Command, error)
	Exports() []string
}

func New(name string, isFinal bool, runOn, failOn string, continueOnFailure bool, factoryCommands commandgroup.CommandGroup, commands commandgroup.AppendableCommandGroup, exportPaths []string) Section {
	if runOn != RunOnAll && runOn != RunOnSplit && runOn != RunOnSingle {
		panic(fmt.Sprintf("Invalid runOn argument: %q", runOn))
	}
	if failOn != FailOnNever && failOn != FailOnAny && failOn != FailOnFirst {
		panic(fmt.Sprintf("Invalid failOn argument: %q", failOn))
	}
	return &section{name, isFinal, runOn, failOn, continueOnFailure, factoryCommands, commands, exportPaths}
}

func (section *section) Name() string {
	return section.name
}

func (section *section) IsFinal() bool {
	return section.isFinal
}

func (section *section) FailOn() string {
	return section.failOn
}

func (section *section) ContinueOnFailure() bool {
	return section.continueOnFailure
}

func (section *section) FactoryCommands(readOnlyCopy bool) commandgroup.CommandGroup {
	if section.factoryCommands == nil {
		return commandgroup.New([]build.Command{})
	}
	if readOnlyCopy {
		return section.factoryCommands.Copy()
	}
	switch section.runOn {
	case RunOnAll, RunOnSplit: // Splits factories in both cases
		return section.factoryCommands
	case RunOnSingle:
		return section.factoryCommands.Remaining()
	}
	panic(fmt.Sprintf("Unexpected runOn value: %s\n", section.runOn))
}

func (section *section) Commands(readOnlyCopy bool) commandgroup.CommandGroup {
	commands := section.commands
	if commands == nil {
		commands = commandgroup.New([]build.Command{})
	}

	if readOnlyCopy {
		return commands.Copy()
	}
	if section.factoryCommands != nil {
		section.factoryCommands.Wait()
	}
	switch section.runOn {
	case RunOnAll:
		return commands.Copy()
	case RunOnSplit:
		return commands
	case RunOnSingle:
		return commands.Remaining()
	}
	panic(fmt.Sprintf("Unexpected runOn value: %s\n", section.runOn))
}

func (section *section) AppendCommand(command build.Command) (build.Command, error) {
	if section.commands == nil {
		section.commands = commandgroup.New([]build.Command{})
	}
	return section.commands.Append(command)
}

func (section *section) Exports() []string {
	return section.exportPaths
}
