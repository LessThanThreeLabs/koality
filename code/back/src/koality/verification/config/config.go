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

const DefaultTimeout = 600

type VerificationConfig struct {
	Params        Params
	Sections      []section.Section
	FinalSections []section.Section
}

type Params struct {
	Nodes          uint
	Environment    map[string]string
	SnapshotUntil  string
	RecursiveClone bool
	GitClean       bool
}

const defaultTimeout = 600

func ParseRemoteCommands(config interface{}, advertised bool) (commands []verification.Command, err error) {
	scripts, ok := config.([]interface{})
	if !ok {
		err = BadConfigurationError{"The scripts subsection should always be a list of either strings or parameter maps."}
		return
	}

	for _, script := range scripts {
		switch script.(type) {
		case string:
			commands = append(commands, remotecommand.NewRemoteCommand(advertised, script.(string), defaultTimeout, nil, []string{script.(string)}))
		case map[interface{}]interface{}:
			script := script.(map[interface{}]interface{})
			for name, parameters := range script {
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

				xunitPaths := []string{}
				timeout := DefaultTimeout
				var command string
				for parameter, value := range paramMap {
					switch parameter {
					case "xunit":
						switch value.(type) {
						case []interface{}:
							paths := value.([]interface{})
							for _, path := range paths {
								path, ok := path.(string)
								if !ok {
									err = BadConfigurationError{"Each xunit path should be a string."}
									return
								}

								xunitPaths = append(xunitPaths, path)
							}
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
							err = BadConfigurationError{"The command parameter of a script should be a string"}
							return
						}
					default:
						err = BadConfigurationError{fmt.Sprintf("The parameter %#v is not supported for scripts.", parameter)}
						return
					}
				}

				if _, ok := paramMap["command"]; !ok {
					command = name
				}

				commands = append(commands, remotecommand.NewRemoteCommand(advertised, name, timeout, xunitPaths, []string{command}))
			}
		default:
			return commands, BadConfigurationError{"The script subsection should contain a list of shell scripts or a map from script names to script parameters."}
		}
	}

	return
}

func parseBool(option interface{}) (val bool, err error) {
	val, ok := option.(bool)
	if !ok {
		return false, BadConfigurationError{}
	}
	return
}

func FromYaml(yamlContents string) (verificationConfig VerificationConfig, err error) {
	var parsedConfig interface{}

	err = goyaml.Unmarshal([]byte(yamlContents), &parsedConfig)
	if err != nil {
		if strings.HasPrefix(err.Error(), "YAML error") {
			return
		} else {
			return
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
			provisionCommand, params, err := convertParameters(config)
			if err != nil {
				return verificationConfig, err
			}

			verificationConfig.Params = params
			provisionShellCommand := verification.NewShellCommand("provision", provisionCommand)
			provisionSection := section.New(
				"provision",
				section.RunOnAll,
				section.FailOnFirst,
				false,
				commandgroup.New([]verification.Command{}),
				commandgroup.New([]verification.Command{provisionShellCommand}),
				[]string{},
			)
			provisionShellCommand = provisionShellCommand
			verificationConfig.Sections = append([]section.Section{provisionSection}, verificationConfig.Sections...)
		case "sections":
			sections, err := parseSections(config)
			if err != nil {
				return verificationConfig, err
			}

			verificationConfig.Sections = append(verificationConfig.Sections, sections...)
		case "final":
			finalSections, err := parseSections(config)
			if err != nil {
				return verificationConfig, err
			}

			verificationConfig.FinalSections = finalSections
		default:
			err = BadConfigurationError{fmt.Sprintf("The primary key %q is not currently supported.", key)}
			return
		}
	}

	return
}

func convertParameters(config interface{}) (provisionCommand shell.Command, params Params, err error) {
	switch config.(type) {
	case map[interface{}]interface{}:
		config := config.(map[interface{}]interface{})
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
				option, ok := option.(uint)
				if !ok || option < 0 {
					err = BadConfigurationError{"The number of nodes must be a positive integer."}
				}

				params.Nodes = uint(option)
			case "snapshot until":
				snapshot, ok := option.(string)
				if !ok {
					err = BadConfigurationError{"The snapshot until parameter has to be a string that identifies another section."}
					return
				}

				params.SnapshotUntil = snapshot
			case "environment":
				variableMap, ok := option.(map[interface{}]interface{})
				if !ok {
					err = BadConfigurationError{"The environment subsection should be a map from environment variables to their desired values."}
					return
				}

				params.Environment = make(map[string]string, len(variableMap))
				for name, value := range variableMap {
					name, ok := name.(string)
					if !ok {
						err = BadConfigurationError{"The keys of the environment map should be strings."}
					}

					params.Environment[name] = fmt.Sprintf("%#v", value)
				}

			case "recursiveClone":
				recursiveClone, err := parseBool(option)
				if err != nil {
					err = BadConfigurationError{"The recursiveClone parameter can be only true or false."}
					return provisionCommand, params, err
				}

				params.RecursiveClone = recursiveClone
			case "gitClean":
				gitClean, err := parseBool(option)
				if err != nil {
					err = BadConfigurationError{"The gitClean parameter can be only true or false."}
					return provisionCommand, params, err
				}

				params.GitClean = gitClean
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

func parseSections(config interface{}) (sections []section.Section, err error) {
	parsedSections, ok := config.([]interface{})
	if !ok {
		err = BadConfigurationError{"Sections should contain a list of sections"}
		return
	}

	for _, section := range parsedSections {
		section, err := parseSection(section)
		if err != nil {
			return sections, err
		}

		sections = append(sections, section)
	}
	return
}

func parseSection(config interface{}) (newSection section.Section, err error) {
	var (
		name, runOn, failOn string
		continueOnFailure   bool
		factoryCommands     commandgroup.CommandGroup
		regularCommands     commandgroup.AppendableCommandGroup
		exportPaths         []string
	)

	singletonMap, ok := config.(map[interface{}]interface{})
	if !ok || len(singletonMap) != 1 {
		err = BadConfigurationError{"Each section should be a map from its name to its parameters."}
		return
	}

	for key := range singletonMap {
		name, ok = key.(string)
		if !ok {
			err = BadConfigurationError{"Each section name should be a string."}
		}
	}

	sectionMap, ok := singletonMap[name].(map[interface{}]interface{})
	if !ok {
		err = BadConfigurationError{"Each section should be a list of scripts and options."}
		return
	}

	for subsection, content := range sectionMap {
		switch subsection {
		case "scripts":
			commands, err := ParseRemoteCommands(content, true)
			if err != nil {
				return newSection, err
			}

			if _, ok := sectionMap["exports"]; ok {
				err = BadConfigurationError{"A section cannot contain both a scripts and an exports subsection."}
				return newSection, err
			}

			regularCommands = commandgroup.New(commands)
		case "factories":
			commands, err := ParseRemoteCommands(content, false)
			if err != nil {
				return newSection, err
			}

			if _, ok := sectionMap["exports"]; ok {
				err = BadConfigurationError{"A section cannot contain both a factories and an exports subsection."}
				return newSection, err
			}

			factoryCommands = commandgroup.New(commands)
		case "run on":
			switch content {
			case "all":
				runOn = section.RunOnAll
			case "split":
				runOn = section.RunOnSplit
			case "single":
				runOn = section.RunOnSingle
			default:
				err = BadConfigurationError{"The run on section should be either all, split or single."}
				return
			}
		case "fail on":
			switch content {
			case "never":
				failOn = section.FailOnNever
			case "any":
				failOn = section.FailOnAny
			case "first":
				failOn = section.FailOnFirst
			default:
				err = BadConfigurationError{"The fail on section should be either never, any or first."}
				return
			}
		case "exports":
			paths, ok := content.([]interface{})
			if !ok {
				err = BadConfigurationError{"The export subsection should be a list of strings."}
				return
			}

			for _, path := range paths {
				path, ok := path.(string)
				if !ok {
					err = BadConfigurationError{"Each export path should be a string."}
					return
				}

				exportPaths = append(exportPaths, path)
			}
		case "continue on failure":
			continueOnFailure, err = parseBool(content)
			if err != nil {
				err = BadConfigurationError{"The recursiveClone parameter can be only true or false."}
				return
			}
		default:
			err = BadConfigurationError{fmt.Sprintf("The %#v parameter is not currently supported for scripts.", subsection)}
		}
	}
	if failOn == "" {
		failOn = section.FailOnAny
	}

	if runOn == "" {
		err = BadConfigurationError{"The run on parameter must be specified for every section."}
		return newSection, err
	}

	if len(exportPaths) == 0 && factoryCommands == nil && regularCommands == nil {
		err = BadConfigurationError{"Each section must have at least one of a scripts, factories or exports subsection."}
		return newSection, err
	}

	return section.New(name, runOn, failOn, continueOnFailure, factoryCommands, regularCommands, exportPaths), nil
}

type BadConfigurationError struct {
	msg string
}

func (e BadConfigurationError) Error() string { return e.msg }
