package stageverifier

import (
	"code.google.com/p/go.crypto/ssh"
	"io"
	"koality/shell"
	"os/exec"
	"sync"
	"syscall"
)

type CommandRunner interface {
	RunCommand(executable shell.Executable) (int, error)
}

type OutputWritingCommandRunner struct {
	writer io.Writer
}

func NewOutputWritingCommandRunner(writer io.Writer) OutputWritingCommandRunner {
	var writeMutex sync.Mutex
	return OutputWritingCommandRunner{&syncWriter{writer, writeMutex}}
}

func (runner OutputWritingCommandRunner) RunCommand(executable shell.Executable) (int, error) {
	var err error

	err = executable.SetStdout(runner.writer)
	if err != nil {
		panic(err)
	}
	err = executable.SetStderr(runner.writer)
	if err != nil {
		panic(err)
	}

	exitErr := executable.Run()
	if exitErr != nil {
		switch exitErr.(type) {
		case *ssh.ExitError:
			sshErr := exitErr.(*ssh.ExitError)
			return sshErr.Waitmsg.ExitStatus(), nil
		case *exec.ExitError:
			execErr := exitErr.(*exec.ExitError)
			// This only works for unix-type systems right now
			waitStatus, ok := execErr.Sys().(syscall.WaitStatus)
			if ok {
				return waitStatus.ExitStatus(), nil
			}
		}
		return -1, exitErr
	}
	return 0, nil
}

type syncWriter struct {
	writer io.Writer
	mutex  sync.Mutex
}

func (writer *syncWriter) Write(bytes []byte) (int, error) {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	return writer.writer.Write(bytes)
}
