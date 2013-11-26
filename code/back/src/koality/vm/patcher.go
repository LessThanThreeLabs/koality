package vm

import (
	"fmt"
	"io/ioutil"
	"koality/shell"
	"os"
)

type Patcher interface {
	Patch(*PatchConfig) (shell.Executable, error)
}

type PatchConfig struct {
	RootDirectory string
	PatchContents string // Reader???????
}

type patcher struct {
	shell.FileCopier
	shell.ExecutableMaker
}

func NewPatcher(fileCopier shell.FileCopier, executableMaker shell.ExecutableMaker) *patcher {
	return &patcher{fileCopier, executableMaker}
}

func (p *patcher) Patch(config *PatchConfig) (shell.Executable, error) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	_, err = tempFile.WriteString(config.PatchContents)
	if err != nil {
		return nil, err
	}
	err = tempFile.Sync()
	if err != nil {
		return nil, err
	}
	cmd, err := p.FileCopier.FileCopy(tempFile.Name(), ".koality_patch")
	// TODO: RUN AND check results
	cmd = cmd
	tempFile.Close()
	os.Remove(tempFile.Name())

	patchCommand := shell.And(
		shell.Command(fmt.Sprintf("echo %s",
			shell.Quote(fmt.Sprintf("%sPATCH CONTENTS:%s",
				shell.AnsiFormat(shell.AnsiFgCyan, shell.AnsiBold),
				shell.AnsiFormat(shell.AnsiReset),
			)),
		)),
		shell.Command("echo"),
		shell.Command("cat ~/.koality_patch"),
		shell.Command("echo"),
		shell.Command(fmt.Sprintf("echo %s",
			shell.Quote(fmt.Sprintf("%sPATCHING:%s",
				shell.AnsiFormat(shell.AnsiFgCyan, shell.AnsiBold),
				shell.AnsiFormat(shell.AnsiReset),
			)),
		)),
		shell.Command("echo"),
		shell.Advertised(shell.Command(fmt.Sprintf("cd %s", config.RootDirectory))),
		shell.Or(
			shell.Advertised(shell.Command("git apply ~/.koality_patch")),
			shell.And(
				shell.Command(fmt.Sprintf("echo -e %s",
					shell.Quote(fmt.Sprintf("%sFailed to git apply, attempting standard patching...%s",
						shell.AnsiFormat(shell.AnsiFgYellow, shell.AnsiBold),
						shell.AnsiFormat(shell.AnsiReset),
					)),
				)),
				shell.Advertised(shell.Command("patch -p1 < ~/.koality_patch")),
			),
		),
	)
	return p.ExecutableMaker.MakeExecutable(patchCommand, nil, nil)
}
