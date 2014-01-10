package config

import (
	"io/ioutil"
	"testing"
)

var buf []byte

func TestErrorFree(test *testing.T) {
	buf, _ = ioutil.ReadFile("example_koality.yml")
	s := string(buf)

	parsedConfig, err := FromYaml(s)

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
