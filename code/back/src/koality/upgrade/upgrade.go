package upgrade

import (
	"fmt"
	"io"
	"koality/util/pathtranslator"
	"os"
	"os/exec"
	"path"
)

func PrepareUpgrade(upgradeTgz io.ReadCloser) (string, error) {
	fmt.Println("Writing upgrade contents...")

	extractCommand := exec.Command("tar", "-xzf", "-")
	extractCommand.Dir = os.TempDir()
	extractCommand.Stdin = upgradeTgz
	if err := extractCommand.Run(); err != nil {
		return "", err
	}

	installerPath := path.Join(os.TempDir(), "out", pathtranslator.BinaryPath("installer"))
	_, err := os.Stat(installerPath)
	if err != nil {
		return "", err
	}
	return installerPath, nil
}

func RunUpgrade(executablePath string) error {
	fmt.Printf("Running installer at %s...\n", executablePath)
	installCommand := exec.Command("sudo", executablePath)
	installCommand.Stdout = os.Stdout
	installCommand.Stderr = os.Stderr
	if err := installCommand.Run(); err != nil {
		return err
	}

	fmt.Println("Restarting...")

	restartCommand := exec.Command("sudo", "service", "koality", "restart")
	return restartCommand.Run()
}
