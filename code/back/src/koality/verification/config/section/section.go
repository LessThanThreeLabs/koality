package section

import (
	"fmt"
	"koality/verification"
	"koality/verification/config/commandgroup"
)

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
	runOn             string
	failOn            string
	continueOnFailure bool
	factoryCommands   commandgroup.CommandGroup
	commands          commandgroup.AppendableCommandGroup
}

type Section interface {
	Name() string
	FailOn() string
	ContinueOnFailure() bool
	FactoryCommands() commandgroup.CommandGroup
	Commands() commandgroup.CommandGroup
	AppendCommand(verification.Command) error
	Exports() []string
}

func New(name, runOn, failOn string, continueOnFailure bool, factoryCommands commandgroup.CommandGroup, commands commandgroup.AppendableCommandGroup) *section {
	if runOn != RunOnAll && runOn != RunOnSplit && runOn != RunOnSingle {
		panic(fmt.Sprintf("Invalid runOn argument: %q", runOn))
	}
	if failOn != FailOnNever && failOn != FailOnAny && failOn != FailOnFirst {
		panic(fmt.Sprintf("Invalid failOn argument: %q", failOn))
	}
	return &section{name, runOn, failOn, continueOnFailure, factoryCommands, commands}
}

func (section *section) Name() string {
	return section.name
}

func (section *section) FailOn() string {
	return section.failOn
}

func (section *section) ContinueOnFailure() bool {
	return section.continueOnFailure
}

func (section *section) FactoryCommands() commandgroup.CommandGroup {
	switch section.runOn {
	case RunOnAll, RunOnSplit: // Splits factories in both cases
		return section.factoryCommands
	case RunOnSingle:
		return section.factoryCommands.Remaining()
	}
	panic(fmt.Sprintf("Unexpected runOn value: %s\n", section.runOn))
}

func (section *section) Commands() commandgroup.CommandGroup {
	section.factoryCommands.Wait()
	switch section.runOn {
	case RunOnAll:
		return section.commands.Copy()
	case RunOnSplit:
		return section.commands
	case RunOnSingle:
		return section.commands.Remaining()
	}
	panic(fmt.Sprintf("Unexpected runOn value: %s\n", section.runOn))
}

func (section *section) AppendCommand(command verification.Command) error {
	return section.commands.Append(command)
}

func (section *section) Exports() []string {
	panic("Not implemented")
}
