package config

import (
	"io/ioutil"
	"testing"
)

var buf []byte

func TestErrorFree(testing *testing.T) {
	buf, _ = ioutil.ReadFile("example_koality.yml")
	s := string(buf)

	_, err := FromYaml(s)

	if err != nil {
		testing.Error(err)
	}
}
