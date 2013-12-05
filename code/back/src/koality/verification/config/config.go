package config

//TODO(documentation and proper error handling)

import (
	"fmt"
	"github.com/dchest/goyaml"
	"koality/shell"
	"koality/verification"
	"koality/verification/config/commandgroup"
	"koality/verification/config/provision"
	"koality/verification/config/remotecommand"
	"koality/verification/config/section"
	"strings"
)

type VerificationConfig struct {
	Params        Params
	Sections      []section.Section
	FinalSections []section.Section
}

type Params struct {
	Nodes          uint
	Snapshot       bool // TODO: should be SnapshotUntil (string)
	RecursiveClone bool
	GitClean       bool
}

const defaultTimeout = 600

//Too long of a function name?
func getRemoteCommandsFromScripts(scripts interface{}, xunit, advertised bool) (commands []verification.TestCommand, err error) {
	switch scripts.(type) {
	case []string:
		commands = append(commands, remotecommand.NewRemoteCommand(true, "setup", defaultTimeout, nil, scripts.([]string)))
		return
	case map[interface{}]interface{}:
		scripts := scripts.(map[interface{}]interface{})
		for name, parameters := range scripts {
			name, ok := name.(string)
			if !ok {
				err = BadConfigurationError{fmt.Sprintf("The name %#v is not a valid string", name)}
				return
			}

			paramMap, ok := parameters.(map[interface{}]interface{})

			if !ok {
				err = BadConfigurationError{fmt.Sprintf("The parameters for script %s should be a map.", name)}
				return
			}

			var xunitPaths []string
			var timeout int
			var command string
			for parameter, value := range paramMap {
				switch parameter {
				case "xunit":
					if !xunit {
						err = BadConfigurationError{"Only the scripts in the test section can have xunit output."}
						return
					}

					switch value.(type) {
					case []string:
						xunitPaths = value.([]string)
					case string:
						xunitPaths = []string{value.(string)}
					default:
						err = BadConfigurationError{"The xunit parameter should be a string or a list of strings."}
						return
					}
				case "timeout":
					timeout, ok = value.(int)

					if !ok {
						err = BadConfigurationError{"The timeout parameter should be an integer."}
						return
					}
				case "command":
					command, ok = value.(string)

					if !ok {
						err = BadConfigurationError{"The command paremeter of a script should be a string"}
						return
					}
				default:
					err = BadConfigurationError{fmt.Sprintf("The parameter %#v is not supported for scripts.", parameter)}
					return
				}
			}

			commands = append(commands, remotecommand.NewRemoteCommand(advertised, name, timeout, xunitPaths, []string{command}))
		}
	default:
		return commands, BadConfigurationError{"The script subsection should contain a list of shell scripts or a map from script names to script parameters."}
	}

	return
}

func getRemoteCommands(scripts interface{}, advertised bool) (commands []verification.Command, err error) {
	remoteCommands, err := getRemoteCommandsFromScripts(scripts, false, advertised)

	// Either this or code duplication. Seriously, does anyone know a better way to do this? Discuss this with me.
	for _, command := range remoteCommands {
		commands = append(commands, command.(verification.Command))
	}
	return
}

func getRemoteTestCommands(scripts interface{}, advertised bool) (commands []verification.TestCommand, err error) {
	return getRemoteCommandsFromScripts(scripts, true, advertised)
}

func FromYaml(yamlContents string) (verificationConfig VerificationConfig, err error) {
	var parsedConfig interface{}

	err = goyaml.Unmarshal([]byte(yamlContents), &parsedConfig)
	if err != nil {
		if strings.HasPrefix(err.Error(), "YAML error") {
			return
		} else {
			// TODO(andrey) log the error (this should be an oom/null-pointer type of error)
		}
	}

	configMap, ok := parsedConfig.(map[interface{}]interface{})

	if !ok {
		err = BadConfigurationError{"The configuration should be a map. For more information go to https://koalitycode.com/documentation."}
		return
	}

	for key, config := range configMap {
		switch key {
		case "parameters":
			nodes, snapshot, recursiveClone, gitClean, provisionCommand, err := convertParameters(config)

			if err != nil {
				return verificationConfig, err
			}

			verificationConfig.Params = Params{nodes, snapshot, recursiveClone, gitClean}
			provisionShellCommand := verification.NewShellCommand("provision", provisionCommand)
			provisionSection := section.New(
				"provision",
				section.RunOnAll,
				section.FailOnFirst,
				false,
				commandgroup.New([]verification.Command{}),
				commandgroup.New([]verification.Command{provisionShellCommand}),
			)
			verificationConfig.Sections = append([]section.Section{provisionSection}, verificationConfig.Sections...)
		case "sections":
			// TODO: this should be filled in from parsing
			sections := []section.Section{}
			verificationConfig.Sections = append(verificationConfig.Sections, sections...)
			// setupCommands, err := convertSetupSection(config)

			// if err != nil {
			// 	return verificationConfig, err
			// }

			// verificationConfig.SetupCommands = setupCommands
		// case "before tests":
		// 	return
		// case "tests":
		// 	return
		// case "after tests":
		// 	return
		default:
			err = BadConfigurationError{fmt.Sprintf("The primary key %q is not currently supported.", key)}
			return
		}
	}

	return
}

func convertParameters(config interface{}) (node uint, snapshot, recursiveClone, gitClean bool, provisionCommand shell.Command, err error) {
	switch config.(type) {
	case map[interface{}]interface{}:
		config := config.(map[interface{}]interface{})

		parseBool := func(option interface{}) (bool, error) {
			switch option {
			case true, false:
				option := option.(bool)
				return option, nil
			default:
				// @Jordan,Brian - what do you think of this? I send this error and catch it and give a more specific one
				return false, BadConfigurationError{}
			}
		}

		for key, option := range config {
			switch key {
			case "languages":

				option, ok := option.(map[interface{}]interface{})

				if !ok {
					err = BadConfigurationError{"Unable to parse language selection.\n" +
						"This should be specified as a mapping from language to version.\n" +
						"Example:\n" +
						"  python: 2.7\n" +
						"  ruby: 1.9.3\n" +
						"See https://koalitycode.com/documentation?view=yamlLanguages for more information."}
					return
				}

				provisionCommand, err = provision.ParseLanguages(option)

				if err != nil {
					return
				}
			case "nodes":
				option, ok := option.(int)

				if !ok || option < 0 {
					err = BadConfigurationError{"The number of nodes must be a positive integer."}
				}

				node = uint(option)
			// Can you guys think of a way to avoid this code duplication?
			case "snapshot":
				snapshot, err = parseBool(option)
				if err != nil {
					err = BadConfigurationError{"The snapshot parameter can be only true or false."}
					return
				}
			case "recursiveClone":
				recursiveClone, err = parseBool(option)
				if err != nil {
					err = BadConfigurationError{"The recursiveClone parameter can be only true or false."}
					return
				}
			case "gitClean":
				gitClean, err = parseBool(option)
				if err != nil {
					err = BadConfigurationError{"The gitClean parameter can be only true or false."}
					return
				}
			default:
				err = BadConfigurationError{fmt.Sprintf("The option %v is not currently supported.", option)}
				return
			}
		}
	default:
		err = BadConfigurationError{"The parameter section should be a map."}
		return
	}
	return
}

func convertSetupSection(config interface{}) (commands []verification.Command, err error) {
	switch config.(type) {
	case map[interface{}]interface{}:
		config := config.(map[interface{}]interface{})
		scripts, ok := config["scripts"]
		if len(config) > 1 {
			return commands, BadConfigurationError{"We currently only support a script subsection for the setup section."}
		} else if len(config) == 1 && !ok {
			return commands, BadConfigurationError{"We currently only support a script subsection for the setup section."}
		} else if ok {
			commands, err = getRemoteCommands(scripts, true)
		}
	default:
		return commands, BadConfigurationError{"The setup section should be a list of scripts"}
	}
	return
}

/*
func convertBeforeTestSection(config interface{}) (firstCommands, everyCommands []verification.Command, err error) {
	switch config.(type) {
	case map[interface{}]interface{}:
		config := config.(map[interface{}]interface{})
		for subsection, content := range config {
			switch subsection {
			case "scripts":
				if len(config) > 1 {
					err = BadConfigurationError{"If you have scripts directly under before tests, then you can have no other subsections."}
				}

			case "first node":

			case "every node":

			default:
			}
		}
	default:
		err = BadConfigurationError{"The setup section should be a list of scripts"}
		return
	}
	return
}
*/

// func convertTestSection(config interface{}) (commands []shell.Command, err error) {
// 	return
// }

// func convertAfterTestSection(config interface{}) (commands []shell.Command, err error) {
// 	return
// }

// func convertDeploySection(config interface{}) (commands []shell.Command, err error) {
// 	return
// }

type BadConfigurationError struct {
	msg string
}

func (e BadConfigurationError) Error() string { return e.msg }
