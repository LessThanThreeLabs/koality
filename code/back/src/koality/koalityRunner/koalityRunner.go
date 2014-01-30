package main

import (
	"fmt"
	"koality/shell"
	"koality/util/pathtranslator"
	"os"
	"os/exec"
	"path"
	"strings"
)

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

	select {
	case err := <-koalityErrorChan:
		panic(fmt.Sprintf("Koality daemon errored out, error: %v", err))
	case err := <-nginxErrorChan:
		panic(fmt.Sprintf("Nginx daemon errored out, error: %v", err))
	}
}

func runDaemon(username, binary string, args ...string) <-chan error {
	errorChan := make(chan error)
	go func(errorChan chan<- error) {
		for {
			shellCommand := shell.AsUser(username, shell.Command(strings.Join(append([]string{binary}, args...), " ")))
			command := exec.Command("bash", "-c", string(shellCommand))
			command.Stdout, command.Stderr = os.Stdout, os.Stderr
			if err := command.Start(); err != nil {
				errorChan <- fmt.Errorf("Could not start daemon %s, error: %v", binary, err)
				return
			}
			if err := command.Wait(); err != nil {
				fmt.Fprintf(os.Stderr, "Daemon %s exited with error: %v\n", binary, err)
			}
		}
	}(errorChan)
	return errorChan
}
