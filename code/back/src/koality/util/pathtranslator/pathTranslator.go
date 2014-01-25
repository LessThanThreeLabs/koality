package pathtranslator

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

func TranslatePath(relativePath string) (string, error) {
	return TranslatePathWithCheckFunc(relativePath, FileExists)
}

func TranslatePathWithCheckFunc(relativePath string, checkFunc func(filePath string) error) (string, error) {
	resolvedPath, err := translatePathFromCurrentExecutable(relativePath, checkFunc)
	if err != nil {
		resolvedPath, err = translatePathFromGoEnvironment(relativePath, checkFunc)
		if err != nil {
			return "", err
		}
	}
	return filepath.Abs(resolvedPath)
}

func translatePathFromCurrentExecutable(relativePath string, checkFunc func(filePath string) error) (string, error) {
	currentBinaryPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	// koality/code/back/bin/executable
	basePath := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentBinaryPath))))
	return resolvePath(basePath, relativePath, checkFunc)
}

func translatePathFromGoEnvironment(relativePath string, checkFunc func(filePath string) error) (string, error) {
	goPathEnv := os.Getenv("GOPATH")
	if goPathEnv == "" {
		return "", errors.New("GOPATH environment variable not set")
	}

	// koality/code/back
	basePath := filepath.Dir(filepath.Dir(goPathEnv))
	return resolvePath(basePath, relativePath, checkFunc)
}

func resolvePath(basePath, relativePath string, checkFunc func(filePath string) error) (string, error) {
	resolvedPath := filepath.Join(basePath, relativePath)

	if err := checkFunc(resolvedPath); err != nil {
		return "", err
	}

	return resolvedPath, nil
}

func BinaryPath(binaryPath string) string {
	return filepath.Join("code", "back", "bin", binaryPath)
}