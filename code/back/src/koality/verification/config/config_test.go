package config

import (
	"fmt"
	"io/ioutil"
	"testing"
)

var buf []byte

func TestErrorFree(test *testing.T) {
	buf, _ = ioutil.ReadFile("example_koality.yml")
	s := string(buf)

	parsedConfig, err := FromYaml(s, ".")

	if err != nil {
		test.Error(err)
	}

	if &(parsedConfig.Params) == nil {
		test.Error("The params section was nil on return.")
	}

	if parsedConfig.Sections == nil || len(parsedConfig.Sections) != 6 {
		test.Errorf("There should be six sections in the example config file, but there are only %d.", len(parsedConfig.Sections))
	}

	if parsedConfig.FinalSections == nil || len(parsedConfig.FinalSections) != 2 {
		test.Errorf("There should be two final sections in the example config file, but there are only %d.", len(parsedConfig.FinalSections))
	}
}

func TestScriptsSubsectionType(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
      chicken chicken:
        command: echo "chicken"
        timeout: 42
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The scripts subsection should be a list")
	}
}

func TestScriptsSubsectionNameValidity(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - 1:
        command: echo "chicken"
        timeout: 42
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The named scripts should have string names")
	}
}

func TestScriptParameterIsMap(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        - command: echo "chicken"
        - timeout: 42
        - xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Named scripts should be a map from parameters to values")
	}
}

func TestScriptXunitPathIsString(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: 42
        xunit:
        - 1
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Xunit paths should be strings")
	}
}

func TestScriptXunitPathsIsList(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: 42
        xunit:
          1`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The xunit parameter should either be a string or a list of strings")
	}
}

func TestScriptTimeoutIsInt(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: not an int
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The timeout parameter should be an integer")
	}
}

func TestScriptParameterIsList(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: true
        timeout: 42
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The command parameter should be a string")
	}
}

func TestScriptParameterIsSupported(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: 42
        chicken: chcken
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("All the parameters of a script should be supported")
	}
}

func TestYamlFileIsMap(test *testing.T) {
	ymlContents := `- sections:
  - setup chicken:
      run on: all
      fail on: first
      scripts:
      - chicken chicken
      - chicken chicken:
          command: echo "chicken"
          timeout: 42
          xunit:
          - chicken chicken
          - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Named scripts should be a map from parameters to values")
	}
}

func TestAllKeysAreValid(test *testing.T) {
	ymlContents := `chicken:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: 42
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The key chicken should probably never be supported")
	}
}

func TestLanguagesIsProper(test *testing.T) {
	ymlContents := `parameters:
  languages:
    - python: 2.7`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The languages section needs to have a specific format")
	}

	ymlContents = `parameters:
  languages:
    chicken: 2.7`

	_, err = FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The languages section needs to have a specific format")
	}

	ymlContents = `parameters:
  languages:
    1: 2.7`

	_, err = FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The languages section needs to have a specific format")
	}

	fmt.Println(err)
}

func TestNodesIsInt(test *testing.T) {
	ymlContents := `parameters:
  nodes: -1`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The nodes subsection needs to be a positive integer")
	}
}

func TestSnapshotUntilIsString(test *testing.T) {
	ymlContents := `parameters:
  snapshot until: 1`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The snapshot until parameter needs to be a string")
	}
}

func TestSnapshotUntilParameterIsValid(test *testing.T) {
	ymlContents := `parameters:
  snapshot until: not setup chicken
sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: 42
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The snapshot until parameter should point to a valid section")
	}

	ymlContents = `parameters:
  snapshot until: setup chicken
sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: 42
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err = FromYaml(ymlContents, ".")

	if err != nil {
		test.Errorf("The snapshot until parameter should point to a valid section")
	}
}

func TestEnvironmentVariables(test *testing.T) {
	ymlContents := `parameters:
  environment:
    - CI: true`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The environment parameter should be a map, not a list")
	}

	ymlContents = `parameters:
  environment:
    1: true`

	_, err = FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The environment parameter's keys should be integers")
	}

	ymlContents = `parameters:
  environment:
    0A: true`

	_, err = FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The environment parameter's keys should conform to the environment variable regexp")
	}
}

func TestRecursiveCloneIsBool(test *testing.T) {
	ymlContents := `parameters:
  recursiveClone: 1`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The recursive clone parameter should be a boolean")
	}
}

func TestGitCleanIsBool(test *testing.T) {
	ymlContents := `parameters:
  gitClean: 1`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The git clean parameter should be a boolean")
	}
}

func TestUnsupportedParameter(test *testing.T) {
	ymlContents := `parameters:
  chicken: chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("All of the paremeters must be supported")
	}
}

func TestParametersIsMap(test *testing.T) {
	ymlContents := `parameters:
  - chicken: chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The parameters section should always be a map")
	}
}

func TestSectionsIsList(test *testing.T) {
	ymlContents := `sections:
setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
    - chicken chicken:
        command: echo "chicken"
        timeout: 42
        xunit:
        - chicken chicken
        - chicken chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The sections key should point ot a list")
	}
}

func TestSectionIsSingletonMap(test *testing.T) {
	ymlContents := `sections:
sections:
- setup chicken:
    run on: all
    fail on: first
    scripts:
    - chicken chicken
  destroy chicken:
    scripts:
    - chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Each section should be a singleton map")
	}
}

func TestSectionNameIsString(test *testing.T) {
	ymlContents := `sections:
- false:
    run on: all
    fail on: first
    scripts:
    - chicken chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Each section should have a string name")
	}
}

func TestSectionIsMap(test *testing.T) {
	ymlContents := `sections:
- setup chicken:
    - run on: all`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Each section should be a map")
	}
}

func TestSectionDoesNotHaveBothScriptsAndExports(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    scripts:
    - chicken
    exports:
    - path/to/chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Test section does not have both scripts and exports")
	}
}

func TestSectionDoesNotHaveBothFactoriesAndExports(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    factories:
    - chicken
    exports:
    - path/to/chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Test section does not have both factories and exports")
	}
}

func TestRunOnIsValid(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    run on: chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Run on should only be alll, split or single")
	}
}

func TestFailOnIsValid(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    fail on: chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Fail on should only be alll, split or single")
	}
}

func TestExportsIsList(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    exports:
      - chicken: chicken`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The exports subsection should be a list")
	}
}

func TestExportPathsAreStrings(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    exports:
      - chicken
      - 1`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("The export paths should all be strings")
	}
}

func TestContinueOnFailureIsBool(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    continue on failure: 1`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("Continue on failure should be a string")
	}
}

func TestSectionParameterIsSupported(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    stuff: stuff`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("All of the parameters of a section are supported")
	}
}

func TestRunOnMustBeSpecified(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    fail on: any`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("A section without a run on parameter is not valid")
	}
}

func TestSectionMustRunSomething(test *testing.T) {
	ymlContents := `sections:
- export chicken:
    run on: all`

	_, err := FromYaml(ymlContents, ".")

	if err == nil {
		test.Errorf("A section must contain at least one of scripts, exports or factories")
	}
}
