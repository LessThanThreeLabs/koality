package config

//TODO(documentation and proper error handling)

import (
	"fmt"
	"github.com/dchest/goyaml"
	"koality/shell"
	"koality/verification"
	"koality/verification/config/provision"
	//"koality/verification/config/remotecommand"
	"strings"
)

/*
type VerificationConfig struct {
	NumMachines     int
	SetupCommands   []verification.Command
	CompileCommands []verification.Command
	FactoryCommands []verification.Command
	TestCommands    []verification.Command
}
*/

type VerificationConfig struct {
	nodes            int
	snapshot         bool
	recursiveClone   bool
	gitClean         bool
	provisionCommand verification.Command
}

func FromYaml(yamlContents string) (verificationConfig VerificationConfig, err error) {
	var parsedConfig interface{}

	err = goyaml.Unmarshal([]byte(yamlContents), &parsedConfig)
	if err != nil {
		if strings.Contains(err.Error(), "YAML error") {
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

	for section, config := range configMap {
		switch section {
		case "parameters":
			nodes, snapshot, recursiveClone, gitClean, provisionCommand, err := convertParameterSection(config)

			if err != nil {
				return verificationConfig, err
			}

			verificationConfig.nodes = nodes
			verificationConfig.snapshot = snapshot
			verificationConfig.recursiveClone = recursiveClone
			verificationConfig.gitClean = gitClean
			verificationConfig.provisionCommand = verification.NewShellCommand("provision", provisionCommand)
		default:
			err = BadConfigurationError{fmt.Sprintf("The section %s is not currently supported.", section)}
			return
		}
	}

	return
}

func convertParameterSection(config interface{}) (node int, snapshot, recursiveClone, gitClean bool, provisionCommand shell.Command, err error) {
	switch config.(type) {
	case map[interface{}]interface{}:
		for key, option := range config.(map[interface{}]interface{}) {
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
			case "node":
				option, ok := option.(int)

				if !ok {
					err = BadConfigurationError{"The number of nodes should be an integer."}
				}

				node = option
			case "snapshot":

			}
		}
	default:
		//TODO (andrey) error
		fmt.Printf("Boo")
	}

	return
}

/*
func convertSetupSection(config interface{}) (command RemoteCommand, err error) {
	switch configType := config.(type) {
	case map[string]interface{}:
		config := config.(map[string]interface{})
		scripts, ok := config["scripts"]
		if len(config) > 1 {
			return commands, BadConfigurationError{"We currently only support a script subsection for the setup section."}
		} else if len(config) == 1 && !ok {
			return commands, BadConfigurationError{"We currently only support a script subsection for the setup section."}
		} else if ok {
			switch scriptType := scripts.(type) {
			case []string:
				commands := scripts.([]string)

			default:
				return commands, BadConfigurationError{"The script subsection should contain a list of shell scripts."}
			}
		}
	default:
		return commands, BadConfigurationError{"The setup section should be a list of scripts"}
	}
	return
}
*/
/*
func convertBeforeTestSection(config interface{}) (commands []shell.Command, err error) {
	switch configType := config.(type) {
	case map[string]interface{}:
		config := config.(map[string]interface{})
		if len(config) > 2 {
			return commands, BadConfigurationError{"We currently only support \"first node\" and \"every node\" subsections for the before tests section."}
		} else if ok {
			switch scriptType = scripts.(type) {
			case []string:
					for _, command := range scripts {
						commands = append(commands, toScript(command))
					}
			default:
				return commands, BadConfigurationError{"The script subsection should contain a list of shell scripts."}
			}
		}
	default:
		return commands, BadConfigurationError{"The setup section should be a list of scripts"}
	}
}
*/

func convertTestSection(config interface{}) (commands []shell.Command, err error) {
	return
}

func convertAfterTestSection(config interface{}) (commands []shell.Command, err error) {
	return
}

func convertDeploySection(config interface{}) (commands []shell.Command, err error) {
	return
}

type BadConfigurationError struct {
	msg string
}

func (e BadConfigurationError) Error() string { return e.msg }
