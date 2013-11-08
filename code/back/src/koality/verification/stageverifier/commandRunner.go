package stageverifier

import (
	"code.google.com/p/go.crypto/ssh"
	"io"
	"koality/shell"
	"os/exec"
	"syscall"
	"time"
)

type CommandRunner interface {
	RunCommand(executable shell.Executable) (int, error)
}

type OutputWritingCommandRunner struct {
	Writer io.Writer
}

func (runner OutputWritingCommandRunner) RunCommand(executable shell.Executable) (int, error) {
	stdoutChan := make(chan []byte)
	stderrChan := make(chan []byte)
	exitChan := make(chan error)

	stdoutPipe, err := executable.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderrPipe, err := executable.StderrPipe()
	if err != nil {
		panic(err)
	}

	executable.Start()

	go runner.handleOutput(stdoutPipe, stdoutChan)
	go runner.handleOutput(stderrPipe, stderrChan)
	go func() {
		exitChan <- executable.Wait()
	}()

	exitErr := runner.processOutput(exitChan, stdoutChan, stderrChan, 100*time.Millisecond)
	if exitErr != nil {
		switch exitErr.(type) {
		case *ssh.ExitError:
			sshErr := exitErr.(*ssh.ExitError)
			return sshErr.Waitmsg.ExitStatus(), nil
		case *exec.ExitError:
			execErr := exitErr.(*exec.ExitError)
			// This only works for unix-type systems right now
			waitStatus, ok = execErr.Sys().(syscall.WaitStatus)
			if ok {
				return waitStatus.ExitStatus(), nil
			}
		}
		return -1, exitErr
	}
	return 0, nil
}

func (runner OutputWritingCommandRunner) handleOutput(reader io.Reader, outChan chan<- []byte) {
	for {
		buffer := make([]byte, 1024)
		numBytes, err := reader.Read(buffer)
		if numBytes > 0 {
			outChan <- buffer[:numBytes]
		}
		if err != nil {
			close(outChan)
			return
		}
	}
}

func (runner OutputWritingCommandRunner) processOutput(exitChan chan error, stdoutChan, stderrChan chan []byte, postExitTimeout time.Duration) error {
	processUntilExit := func() error {
		for {
			select {
			case err := <-exitChan:
				return err
			case out, ok := <-stdoutChan:
				if out != nil {
					runner.Writer.Write(out)
				}
				if !ok {
					stdoutChan = nil
				}
			case out, ok := <-stderrChan:
				if out != nil {
					runner.Writer.Write(out)
				}
				if !ok {
					stderrChan = nil
				}
			}
		}
	}

	processRemainingOutput := func() {
		timeout := time.After(postExitTimeout)
		for {
			if stdoutChan == nil && stderrChan == nil {
				return
			}

			select {
			case <-timeout:
				return
			case out, ok := <-stdoutChan:
				if out != nil {
					runner.Writer.Write(out)
				}
				if !ok {
					stdoutChan = nil
				}
			case out, ok := <-stderrChan:
				if out != nil {
					runner.Writer.Write(out)
				}
				if !ok {
					stderrChan = nil
				}
			}
		}
	}

	exitErr := processUntilExit()
	processRemainingOutput()
	return exitErr
}
