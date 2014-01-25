package pathtranslator

import (
	"fmt"
	"os"
)

// Default file check method, used by TranslatePath
func CheckExists(filePath string) error {
	_, err := os.Stat(filePath)
	return err
}

func CheckExecutable(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	fileMode := fileInfo.Mode()
	if !fileMode.IsRegular() {
		return fmt.Errorf("File found at %s is not a regular file", filePath)
	}

	if fileMode.Perm()&0500 == 0 {
		return fmt.Errorf("File found at %s does not have read and execute permissions", filePath)
	}

	return nil
}
