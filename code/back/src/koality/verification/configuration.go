package verification

type VerificationConfig struct {
	numMachines     int
	setupCommands   []command
	compileCommands []command
	factoryCommands []command
	testCommands    []command
}

type result struct {
	stageType string
	passed    bool
}

type command struct{}
