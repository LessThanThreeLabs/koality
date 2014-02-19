package main

import (
	"fmt"
	"koality/shell"
	"koality/util/pathtranslator"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

var daemons = make(map[string]*exec.Cmd, 2)
var exited = false

func main() {
	koalityBinary, err := pathtranslator.TranslatePathWithCheckFunc(pathtranslator.BinaryPath("koality"), pathtranslator.CheckExecutable)
	if err != nil {
		panic("Could not find the koality binary")
	}

	nginxBinary, err := pathtranslator.TranslatePathWithCheckFunc(path.Join("nginx", "nginx"), pathtranslator.CheckExecutable)
	if err != nil {
		panic("Could not find the nginx binary")
	}

	koalityErrorChan := runDaemon("koality", koalityBinary)
	nginxErrorChan := runDaemon("root", nginxBinary, "-c", path.Join(path.Dir(nginxBinary), "nginx.conf"))
	signalChan := make(chan os.Signal, 2)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-koalityErrorChan:
		panic(fmt.Sprintf("Koality daemon errored out, error: %v", err))
	case err := <-nginxErrorChan:
		panic(fmt.Sprintf("Nginx daemon errored out, error: %v", err))
	case signal := <-signalChan:
		exit(signal)
	}
}

func runDaemon(username, binary string, args ...string) <-chan error {
	errorChan := make(chan error)
	go func(errorChan chan<- error) {
		startTime := time.Now()
		for {
			if exited {
				return
			}
			shellCommand := shell.AsUser(username, shell.Command(strings.Join(append([]string{binary}, args...), " ")))
			command := exec.Command("bash", "-c", string(shellCommand))
			daemons[binary] = command
			command.Stdout, command.Stderr = os.Stdout, os.Stderr
			if err := command.Start(); err != nil {
				errorChan <- fmt.Errorf("Could not start daemon %s, error: %v", binary, err)
				return
			}
			startTime = time.Now()
			if err := command.Wait(); err != nil {
				fmt.Fprintf(os.Stderr, "Daemon %s exited with error: %v\n", binary, err)
			}
			runDuration := time.Now().Sub(startTime)
			if runDuration < 5*time.Second {
				fmt.Fprintf(os.Stderr, "Daemon %s exited in under 5 seconds\n", binary)
				time.Sleep(10*time.Second - runDuration)
			}
		}
	}(errorChan)
	return errorChan
}

func exit(signal os.Signal) {
	exited = true
	for _, daemon := range daemons {
		daemon.Process.Signal(syscall.SIGTERM)
		daemon.Process.Wait()
	}
	panic(fmt.Sprintf("Received signal %s, exiting now.", signal.String()))
}
