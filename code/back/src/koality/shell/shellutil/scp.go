package shellutil

import (
	"bytes"
	"fmt"
	"io"
	"koality/shell"
	"os"
	"path"
)

func CreateScpExecutable(localFilePath, remoteFilePath string) (shell.Executable, error) {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return shell.Executable{}, err
	}
	// TODO (bbland): figure out what to do with this file handle
	// defer localFile.Close()

	fileInfo, err := localFile.Stat()
	if err != nil {
		return shell.Executable{}, err
	}
	headerBuffer := bytes.NewBufferString(fmt.Sprintf("C%#o %d %s\n", fileInfo.Mode()&os.ModePerm, fileInfo.Size(), path.Base(remoteFilePath)))
	scpStdin := io.MultiReader(headerBuffer, localFile, bytes.NewReader([]byte{0}))

	// Note: this is more powerful than standard scp, as it will actually create the destination directory for you
	scpCommand := shell.And(
		shell.Commandf("mkdir -p %s", path.Dir(remoteFilePath)),
		shell.Commandf("scp -qrt %s", path.Dir(remoteFilePath)),
	)

	return shell.Executable{
		Command: scpCommand,
		Stdin:   scpStdin,
	}, nil
}
