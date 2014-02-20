package config

//TODO(documentation and proper error handling)

import (
	"fmt"
	"github.com/dchest/goyaml"
	"koality/build"
	"koality/build/config/commandgroup"
	"koality/build/config/provision"
	"koality/build/config/remotecommand"
	"koality/build/config/section"
	"koality/shell"
	"koality/util/log"
	"regexp"
	"strconv"
	"strings"
)

const DefaultTimeout = 600

type BuildConfig struct {
	Params        Params
	Sections      []section.Section
	FinalSections []section.Section
}

type Params struct {
	PoolId         uint64
	Nodes          uint64
	Environment    map[string]string
	SnapshotUntil  string
	RecursiveClone bool
	GitClean       bool
}

const defaultTimeout = 600

var environmentVariableNameRegexp = regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$")

func ParseRemoteCommands(config interface{}, advertised bool, directory string) (commands []build.Command, err error) {
	scripts, ok := config.([]interface{})
	if !ok {
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n %sThe scripts subsection should always be a list of strings and parameter maps, but was of type %T.%s",
			config,
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			scripts,
			shell.AnsiFormat(shell.AnsiReset))}
		return
	}

	for _, script := range scripts {
		switch script.(type) {
		case string:
			commands = append(commands, remotecommand.NewRemoteCommand(advertised, directory, script.(string), defaultTimeout, nil, []string{script.(string)}))
		case map[interface{}]interface{}:
			script := script.(map[interface{}]interface{})
			for name, parameters := range script {
				nameString, ok := name.(string)
				if !ok {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe script name %#v is not a valid string%s",
						script,
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						name,
						shell.AnsiFormat(shell.AnsiReset))}
					return
				}

				paramMap, ok := parameters.(map[interface{}]interface{})
				if !ok {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe parameters for script %q should be a map, but  %v was of type %T%s.",
						script,
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						nameString,
						parameters,
						parameters,
						shell.AnsiFormat(shell.AnsiReset))}
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
								xunitPath, ok := path.(string)
								if !ok {
									err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sXunit path should be a string, but %v was of type %T.%s",
										script,
										shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
										path,
										path,
										shell.AnsiFormat(shell.AnsiReset))}
									return
								}

								xunitPaths = append(xunitPaths, xunitPath)
							}
						case string:
							xunitPaths = []string{value.(string)}
						default:
							err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe xunit parameter should be a string or a list of strings, but %v was %T.%s",
								paramMap,
								shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
								value,
								value,
								shell.AnsiFormat(shell.AnsiReset))}
							return
						}
					case "timeout":
						timeout, ok = value.(int)

						if !ok {
							err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe timeout parameter should be an integer, but %v was of type %T.%s",
								paramMap,
								shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
								value,
								value,
								shell.AnsiFormat(shell.AnsiReset))}
							return
						}
					case "command":
						command, ok = value.(string)
						if !ok {
							err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe command parameter of a script should be a string, but %v was of type %T.%s",
								paramMap,
								shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
								value,
								value,
								shell.AnsiFormat(shell.AnsiReset))}
							return
						}
					default:
						err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe parameter %v is not currently supported for scripts.%s",
							paramMap,
							shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
							parameter,
							shell.AnsiFormat(shell.AnsiReset))}
						return
					}
				}

				if _, ok := paramMap["command"]; !ok {
					command = nameString
				}

				commands = append(commands, remotecommand.NewRemoteCommand(advertised, directory, nameString, timeout, xunitPaths, []string{command}))
			}
		default:
			return commands, BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe scripts subsection should contain a list shell scripts and maps from script names to script parameters, but %v was of type %T.%s",
				scripts,
				shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
				script,
				script,
				shell.AnsiFormat(shell.AnsiReset))}
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

func FromYaml(yamlContents, directory string) (buildConfig BuildConfig, err error) {
	var parsedConfig interface{}

	err = goyaml.Unmarshal([]byte(yamlContents), &parsedConfig)
	if err != nil {
		if strings.HasPrefix(err.Error(), "YAML error") {
			err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%s%s%s",
				shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
				err.Error(),
				shell.AnsiFormat(shell.AnsiReset))}
			return
		} else {
			log.Criticalf("Failure when unmarshaling yaml file (error is %v)", err)
			return
		}
	}

	configMap, ok := parsedConfig.(map[interface{}]interface{})

	if !ok {
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe configuration should be a map, but was a %T. For more information go to https://koalitycode.com/documentation.%s",
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			parsedConfig,
			shell.AnsiFormat(shell.AnsiReset))}
		return
	}

	for key, config := range configMap {
		switch key {
		case "parameters":
			provisionCommand, params, err := convertParameters(config)
			if err != nil {
				return buildConfig, err
			}

			buildConfig.Params = params

			if provisionCommand != nil {
				provisionShellCommand := build.NewShellCommand("provision", *provisionCommand)
				provisionSection := section.New(
					"provision",
					false,
					section.RunOnAll,
					section.FailOnFirst,
					false,
					commandgroup.New([]build.Command{}),
					commandgroup.New([]build.Command{provisionShellCommand}),
					[]string{},
				)
				provisionShellCommand = provisionShellCommand
				buildConfig.Sections = append([]section.Section{provisionSection}, buildConfig.Sections...)
			}
		case "sections":
			sections, err := parseSections(config, directory, false)
			if err != nil {
				return buildConfig, err
			}

			buildConfig.Sections = append(buildConfig.Sections, sections...)
		case "final":
			finalSections, err := parseSections(config, directory, true)
			if err != nil {
				return buildConfig, err
			}

			buildConfig.FinalSections = finalSections
		default:
			err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe primary key %q is not currently supported.%s",
				shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
				key,
				shell.AnsiFormat(shell.AnsiReset))}
			return
		}
	}

	// This check is here, because the order of the configuration keys (parameters, sections and final) is not enforced or guaranteed.
	if buildConfig.Params.SnapshotUntil != "" {
		ok := false

		for _, section := range buildConfig.Sections {
			if section.Name() == buildConfig.Params.SnapshotUntil {
				ok = true
				break
			}
		}

		if !ok {
			return buildConfig, BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe snapshot until parameter has to identify another section, but this configuration file does not specify a section with the name %q.%s",
				shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
				buildConfig.Params.SnapshotUntil,
				shell.AnsiFormat(shell.AnsiReset))}
		}
	}

	return
}

func convertParameters(config interface{}) (provisionCommand *shell.Command, params Params, err error) {
	// TODO(akostov): handle pool id/name defaults to 1, should be default pool
	params.PoolId = 1

	switch config.(type) {
	case map[interface{}]interface{}:
		config := config.(map[interface{}]interface{})
		for key, option := range config {
			switch key {
			case "languages":
				option, ok := option.(map[interface{}]interface{})
				if !ok {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sUnable to parse the languages section.%s\n"+
						"The language section should be specified as a mapping from language to version.\n"+
						"Example:\n\n"+
						"languages:\n"+
						"  python: 2.7\n"+
						"  ruby: 1.9.3\n\n"+
						"See https://koalitycode.com/documentation?view=yamlLanguages for more information.",
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						shell.AnsiFormat(shell.AnsiReset))}
					return
				}

				provisionCommand, err = provision.ParseLanguages(option)

				if err != nil {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%s%s%s",
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						err.Error(),
						shell.AnsiFormat(shell.AnsiReset))}
				}
			case "nodes":
				var intVal int
				intVal, err = strconv.Atoi(fmt.Sprint(option))
				if err != nil || intVal < 0 {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe number of nodes must be a positive integer.%s",
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						shell.AnsiFormat(shell.AnsiReset))}
					return
				}

				params.Nodes = uint64(intVal)
			case "snapshot until":
				snapshot, ok := option.(string)
				if !ok {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe snapshot until parameter has to be a string, but was instead a %T.%s",
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						option,
						shell.AnsiFormat(shell.AnsiReset))}
					return
				}

				params.SnapshotUntil = snapshot
			case "environment":
				variableMap, ok := option.(map[interface{}]interface{})
				if !ok {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe environment subsection should be a map from environment variables to their desired values., but was instead of type %T.%s",
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						option,
						shell.AnsiFormat(shell.AnsiReset))}
					return
				}

				params.Environment = make(map[string]string, len(variableMap))
				for name, value := range variableMap {
					variableName, ok := name.(string)
					if !ok {
						err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe keys of the environment map should be strings, but instead %v was %T.%s",
							shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
							name,
							name,
							shell.AnsiFormat(shell.AnsiReset))}
						return
					}

					isValid := environmentVariableNameRegexp.MatchString(variableName)
					if !isValid {
						err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe environment variable %q does not match the regular expression %q.%s",
							shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
							variableName,
							environmentVariableNameRegexp.String(),
							shell.AnsiFormat(shell.AnsiReset))}
						return
					}

					params.Environment[variableName] = fmt.Sprint(value)
				}

			case "recursiveClone":
				recursiveClone, err := parseBool(option)
				if err != nil {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe recursiveClone parameter should only be true or false.%s",
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						shell.AnsiFormat(shell.AnsiReset))}
					return provisionCommand, params, err
				}

				params.RecursiveClone = recursiveClone
			case "gitClean":
				gitClean, err := parseBool(option)
				if err != nil {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe gitClean parameter should only be true or false.%s",
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						shell.AnsiFormat(shell.AnsiReset))}
					return provisionCommand, params, err
				}

				params.GitClean = gitClean
			default:
				err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe option %s is not currently a supported parameter.%s",
					shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
					option,
					shell.AnsiFormat(shell.AnsiReset))}
				return
			}
		}
	default:
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe parameter section should be a map, but instead was of type %T.%s",
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			config,
			shell.AnsiFormat(shell.AnsiReset))}
		return
	}
	return
}

func parseSections(config interface{}, directory string, final bool) (sections []section.Section, err error) {
	parsedSections, ok := config.([]interface{})
	if !ok {
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing the yaml file.\n%sThe sections key should contain a list of sections.%s",
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			shell.AnsiFormat(shell.AnsiReset))}
		return
	}

	for _, section := range parsedSections {
		section, err := parseSection(section, directory, final)
		if err != nil {
			return sections, err
		}

		sections = append(sections, section)
	}
	return
}

func parseSection(config interface{}, directory string, final bool) (newSection section.Section, err error) {
	var (
		name, runOn, failOn string
		continueOnFailure   bool
		factoryCommands     commandgroup.CommandGroup
		regularCommands     commandgroup.AppendableCommandGroup
		exportPaths         []string
	)

	singletonMap, ok := config.(map[interface{}]interface{})
	if !ok || len(singletonMap) != 1 {
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sEach section should be a map from its name to its parameters, but %q was not.%s",
			config,
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			config,
			shell.AnsiFormat(shell.AnsiReset))}
		return
	}

	for key := range singletonMap {
		name, ok = key.(string)
		if !ok {
			err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sEach section name should be a string, but %v was of type %T.%s",
				config,
				shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
				key,
				key,
				shell.AnsiFormat(shell.AnsiReset))}
			return
		}
	}

	sectionMap, ok := singletonMap[name].(map[interface{}]interface{})
	if !ok {
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sEach section should be a list of scripts and options, but instead %v was of type %T.%s",
			config,
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			singletonMap[name],
			singletonMap[name],
			shell.AnsiFormat(shell.AnsiReset))}
		return
	}

	for subsection, content := range sectionMap {
		switch subsection {
		case "scripts":
			commands, err := ParseRemoteCommands(content, true, directory)
			if err != nil {
				return newSection, err
			}

			if _, ok := sectionMap["exports"]; ok {
				err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sA section cannot contain both a scripts and an exports subsection.%s",
					config,
					shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
					shell.AnsiFormat(shell.AnsiReset))}
				return newSection, err
			}

			regularCommands = commandgroup.New(commands)
		case "factories":
			commands, err := ParseRemoteCommands(content, false, directory)
			if err != nil {
				return newSection, err
			}

			if _, ok := sectionMap["exports"]; ok {
				err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sA section cannot contain both a factories and an exports subsection.%s",
					config,
					shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
					shell.AnsiFormat(shell.AnsiReset))}
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
				err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe run on section should be either all, split or single.%s",
					config,
					shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
					shell.AnsiFormat(shell.AnsiReset))}
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
				err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe fail on section should be either never, any or first.%s",
					config,
					shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
					shell.AnsiFormat(shell.AnsiReset))}
				return
			}
		case "exports":
			paths, ok := content.([]interface{})
			if !ok {
				err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe export subsection should be a list, but instead %v was of type %T.%s",
					config,
					shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
					content,
					content,
					shell.AnsiFormat(shell.AnsiReset))}
				return
			}

			for _, path := range paths {
				exportPath, ok := path.(string)
				if !ok {
					err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sEach export path should be a string, but %v was of type %T.%s",
						config,
						shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
						path,
						path,
						shell.AnsiFormat(shell.AnsiReset))}
					return
				}

				exportPaths = append(exportPaths, exportPath)
			}
		case "continue on failure":
			continueOnFailure, err = parseBool(content)
			if err != nil {
				err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe continue on failure parameter should only be true or false.%s",
					config,
					shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
					shell.AnsiFormat(shell.AnsiReset))}
				return
			}
		default:
			err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe parameter %v is not currently supported for sections.%s",
				config,
				shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
				subsection,
				shell.AnsiFormat(shell.AnsiReset))}
			return
		}
	}
	if failOn == "" {
		failOn = section.FailOnAny
	}

	if runOn == "" {
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sThe run on parameter must be specified for every section.%s",
			config,
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			shell.AnsiFormat(shell.AnsiReset))}
		return newSection, err
	}

	if len(exportPaths) == 0 && factoryCommands == nil && regularCommands == nil {
		err = BadConfigurationError{fmt.Sprintf("ERROR when parsing %v from the yaml file.\n%sEach section must have at least one of a scripts, factories or exports subsection.%s",
			config,
			shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
			shell.AnsiFormat(shell.AnsiReset))}
		return newSection, err
	}

	return section.New(name, final, runOn, failOn, continueOnFailure, factoryCommands, regularCommands, exportPaths), nil
}

type BadConfigurationError struct {
	msg string
}

func (e BadConfigurationError) Error() string { return e.msg }
