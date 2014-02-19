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
	extractCommand := exec.Command("tar", "-xzf", "-")
	extractCommand.Dir = os.TempDir()
	extractCommand.Stdin = upgradeTgz
	if err := extractCommand.Run(); err != nil {
		return "", err
	}

	installerPath := path.Join(os.TempDir(), "koality", pathtranslator.BinaryPath("installer"))
	_, err := os.Stat(installerPath)
	if err != nil {
		return "", err
	}
	return installerPath, nil
}

func RunUpgrade(executablePath string) error {
	upgradeLogPath := path.Join(os.TempDir(), "koality-upgrade.log")
	upgradeLog, err := os.OpenFile(upgradeLogPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	fmt.Fprintf(upgradeLog, "Running installer at %s...\n", executablePath)
	installCommand := exec.Command("sudo", executablePath)
	installCommand.Stdout = upgradeLog
	installCommand.Stderr = upgradeLog
	if err := installCommand.Run(); err != nil {
		return err
	}

	fmt.Fprintln(upgradeLog, "Upgrade complete, restarting...")

	restartCommand := exec.Command("sudo", "service", "koality", "restart")
	restartCommand.Stdout = upgradeLog
	restartCommand.Stderr = upgradeLog
	return restartCommand.Run()
}
