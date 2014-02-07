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
	executable := shell.Executable{command, stdin, outBuffer, outBuffer, nil}
	execution, err := vm.Execute(executable)
	if err != nil {
		test.Logf("Failed to execute command:\n%s\n", command)
		test.Fatal(err)
	}
	if err = execution.Wait(); err != nil {
		test.Fatal(err)
	}

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

	executable := shell.Executable{command, nil, stdoutBuffer, stderrBuffer, nil}
	execution, err := vm.Execute(executable)
	if err != nil {
		test.Logf("Failed to execute command:\n%s\n", command)
		test.Fatal(err)
	}
	if err = execution.Wait(); err != nil {
		test.Fatal(err)
	}

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

	executable := shell.Executable{command, nil, buffer, nil, nil}
	execution, err := vm.Execute(executable)
	if err != nil {
		test.Logf("Failed to execute command:\n%s\n", command)
		test.Fatal(err)
	}
	if err = execution.Wait(); err != nil {
		test.Fatal(err)
	}

	userVar := strings.TrimSpace(buffer.String())

	if os.Getenv("USER") != userVar {
		test.Errorf("Expected $USER=%s, was %s", os.Getenv("USER"), userVar)
	}

	buffer.Reset()

	fakeUserVar := "one two three"

	executable = shell.Executable{command, nil, buffer, nil, map[string]string{"USER": fakeUserVar}}
	execution, err = vm.Execute(executable)
	if err != nil {
		test.Logf("Failed to execute command:\n%s\n", command)
		test.Fatal(err)
	}
	if err = execution.Wait(); err != nil {
		test.Fatal(err)
	}

	userVar = strings.TrimSpace(buffer.String())

	if fakeUserVar != userVar {
		test.Errorf("Failed to override environment variable.\nExpected $USER=%s, was %s", fakeUserVar, userVar)
	}
}
