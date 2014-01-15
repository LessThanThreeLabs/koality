package find

import (
	"os"
	"testing"
)

func TestFollowSymlink(test *testing.T) {
	var filePaths []string
	userWalkFunc := func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.Mode().IsRegular() {
			filePaths = append(filePaths, path)
			return nil
		}
		return nil
	}
	err := Find("testdir", "*.xml", userWalkFunc)
	if err != nil {
		test.Fatal(err)
	}

	if len(filePaths) != 1 {
		test.Fatal("wrong number of normal files found")
	}

	if filePaths[0] == "sample1.xml" {
		test.Fatal("found the wrong file", filePaths[0])
	}
}
