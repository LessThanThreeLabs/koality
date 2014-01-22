package shell_test

import (
	"bytes"
	"koality/shell"
	"koality/vm/localmachine"
	"os"
	"strings"
	"testing"
)

func TestInputOuputStreams(test *testing.T) {
	vm, err := localmachine.New()
	if err != nil {
		test.Fatal(err)
	}

	defer vm.Terminate()

	inputString := "hello world"
	stdin := bytes.NewBufferString(inputString)
	outBuffer := new(bytes.Buffer)

	command := shell.Command("cat")

	executable, err := vm.MakeExecutable(command, stdin, outBuffer, outBuffer, nil)
	if err != nil {
		test.Logf("Failed to create executable from command:\n%s\n", command)
		test.Fatal(err)
	}
	executable.Run()

	outputString := strings.TrimSpace(outBuffer.String())
	if outputString != inputString {
		test.Errorf("Expected output=%q, was %q", inputString, outputString)
	}
}

func TestStdoutStderrStreams(test *testing.T) {
	vm, err := localmachine.New()
	if err != nil {
		test.Fatal(err)
	}

	defer vm.Terminate()

	stdoutString := "this IS some\nstdout..."
	stderrString := "This shouldn't be in stdout.\n      \t      It is stderr."

	stdoutBuffer := new(bytes.Buffer)
	stderrBuffer := new(bytes.Buffer)

	command := shell.And(
		shell.Commandf("echo %s", shell.Quote(stdoutString)),
		shell.Redirect(shell.Commandf("echo %s", shell.Quote(stderrString)), "/dev/stderr", false),
	)

	executable, err := vm.MakeExecutable(command, nil, stdoutBuffer, stderrBuffer, nil)
	if err != nil {
		test.Logf("Failed to create executable from command:\n%s\n", command)
		test.Fatal(err)
	}
	executable.Run()

	stdout := strings.TrimSpace(stdoutBuffer.String())
	stderr := strings.TrimSpace(stderrBuffer.String())

	if stdout != stdoutString {
		test.Errorf("Expected stdout=%q, was %q", stdoutString, stdout)
	}

	if stderr != stderrString {
		test.Errorf("Expected stderr=%q, was %q", stderrString, stderr)
	}
}

func TestEnvironment(test *testing.T) {
	vm, err := localmachine.New()
	if err != nil {
		test.Fatal(err)
	}

	defer vm.Terminate()

	buffer := new(bytes.Buffer)

	command := shell.Command("echo $USER")

	executable, err := vm.MakeExecutable(command, nil, buffer, nil, nil)
	if err != nil {
		test.Logf("Failed to create executable from command:\n%s\n", command)
		test.Fatal(err)
	}
	err = executable.Run()
	if err != nil {
		test.Logf("Expected command to pass:\n%s\n", command)
		test.Fatal(err)
	}

	userVar := strings.TrimSpace(buffer.String())

	if os.Getenv("USER") != userVar {
		test.Errorf("Expected $USER=%s, was %s", os.Getenv("USER"), userVar)
	}

	buffer.Reset()

	fakeUserVar := "one two three"

	executable, err = vm.MakeExecutable(command, nil, buffer, nil, map[string]string{"USER": fakeUserVar})
	if err != nil {
		test.Logf("Failed to create executable from command:\n%s\n", command)
		test.Fatal(err)
	}
	err = executable.Run()
	if err != nil {
		test.Logf("Expected command to pass:\n%s\n", command)
		test.Fatal(err)
	}

	userVar = strings.TrimSpace(buffer.String())

	if fakeUserVar != userVar {
		test.Errorf("Failed to override environment variable.\nExpected $USER=%s, was %s", fakeUserVar, userVar)
	}
}
