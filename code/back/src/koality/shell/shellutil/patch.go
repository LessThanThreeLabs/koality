package shellutil

import (
	"fmt"
	"io"
	"koality/shell"
)

func CreatePatchExecutable(repositoryDirectory string, patchContents io.Reader) shell.Executable {
	patchCommand := shell.And(
		shell.Commandf("patchfile=%s", shell.Capture("mktemp")),
		shell.Redirect("cat", "\"$patchfile\"", false),
		shell.Commandf("echo -e %s",
			shell.Quote(fmt.Sprintf("%sPATCH CONTENTS:%s",
				shell.AnsiFormat(shell.AnsiFgCyan, shell.AnsiBold),
				shell.AnsiFormat(shell.AnsiReset),
			)),
		),
		shell.Command("echo"),
		shell.Command("cat \"$patchfile\""),
		shell.Command("echo"),
		shell.Commandf("echo -e %s",
			shell.Quote(fmt.Sprintf("%sPATCHING:%s",
				shell.AnsiFormat(shell.AnsiFgCyan, shell.AnsiBold),
				shell.AnsiFormat(shell.AnsiReset),
			)),
		),
		shell.Command("echo"),
		shell.Advertised(shell.Command(fmt.Sprintf("cd %s", repositoryDirectory))),
		shell.Or(
			shell.Advertised(shell.Command("git apply \"$patchfile\"")),
			shell.And(
				shell.Commandf("echo -e %s",
					shell.Quote(fmt.Sprintf("%sFailed to git apply, attempting standard patching...%s",
						shell.AnsiFormat(shell.AnsiFgYellow, shell.AnsiBold),
						shell.AnsiFormat(shell.AnsiReset),
					)),
				),
				shell.Advertised(shell.Command("patch -p1 < \"$patchfile\"")),
			),
		),
		shell.Command("rm \"$patchfile\""),
	)
	return shell.Executable{
		Command: patchCommand,
		Stdin:   patchContents,
	}
}
