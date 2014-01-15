package find

import (
	"os"
	"path/filepath"
)

func Find(root string, pattern string, userWalkFunc filepath.WalkFunc) error {
	walkFunc := func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}

			return Find(realPath, pattern, userWalkFunc)
		}
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil { // pattern is invalid
			return err
		}

		if matched {
			return userWalkFunc(path, fileInfo, err)
		}
		return nil
	}
	return filepath.Walk(root, walkFunc)
}
