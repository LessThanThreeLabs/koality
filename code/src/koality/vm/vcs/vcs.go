package vcs

import (
	"fmt"
	"koality/shell"
)

type VcsType string

type VcsConfig struct {
	RepositoryName string
	RepositoryHost string
	RepositoryPath string
}

type Vcs interface {
	Type() VcsType
	CloneCommand(VcsConfig) shell.Command
	CheckoutCommand(VcsConfig, string) shell.Command
}

func hostAccessCheckCommand(hostUrl string) shell.Command {
	failureMessage := fmt.Sprintf("%sFailed to access the master instance. Please check to make sure your security groups are configured correctly.%s",
		shell.AnsiFormat(shell.AnsiFgYellow, shell.AnsiBold),
		shell.AnsiFormat(shell.AnsiReset),
	)
	return shell.And(
		shell.Command("echo Testing ssh connection to master instance..."),
		shell.Or(
			shell.Advertised(shell.Command(fmt.Sprintf("ssh %s true", hostUrl))),
			shell.And(
				shell.Command(fmt.Sprintf("echo -e %s", shell.Quote(failureMessage))),
				shell.Command("false"),
			),
		),
	)
}

var (
	git = Git{}
	hg  = Hg{}
)

var VcsMap = map[VcsType]Vcs{
	git.Type(): git,
	hg.Type():  hg,
}
