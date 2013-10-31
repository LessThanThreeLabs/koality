package config

import (
	"koality/verification"
)

type VerificationConfig struct {
	NumMachines     int
	SetupCommands   []verification.Command
	CompileCommands []verification.Command
	FactoryCommands []verification.Command
	TestCommands    []verification.Command
}
