package vcs

import (
	"fmt"
	"koality/shell"
)

type Hg struct{} // Empty struct?

func (hg Hg) Type() VcsType {
	return VcsType("hg")
}

func (hg Hg) CloneCommand(vcsConfig VcsConfig) shell.Command {
	repositoryUrl := fmt.Sprintf("%s/%s", vcsConfig.RepositoryHost, vcsConfig.RepositoryPath)
	cloneCommand := shell.And(
		shell.Or(
			shell.Not(shell.Test(shell.Command(fmt.Sprintf("-e %s", vcsConfig.RepositoryName)))),
			shell.Advertised(shell.Command(fmt.Sprintf("rm -rf %s", vcsConfig.RepositoryName))),
		),
		hostAccessCheckCommand(vcsConfig.RepositoryHost),
		shell.Advertised(shell.Command(fmt.Sprintf("hg clone --uncompressed %s %s", repositoryUrl, vcsConfig.RepositoryName))),
	)
	return cloneCommand
}

func (hg Hg) CheckoutCommand(vcsConfig VcsConfig, ref string) shell.Command {
	repositoryUrl := fmt.Sprintf("%s/%s", vcsConfig.RepositoryHost, vcsConfig.RepositoryPath)
	checkoutCommand := shell.And(
		hostAccessCheckCommand(vcsConfig.RepositoryHost),
		shell.Or(
			shell.Silent(shell.Command("which hg")),
			shell.Sudo(shell.Command("apt-get install -y mercurial")), // TODO: make this not just apt
		),
		shell.IfElse(
			shell.Silent(shell.Command(fmt.Sprintf("mv /repositories/cached/%s %s", vcsConfig.RepositoryName, vcsConfig.RepositoryName))),
			shell.And(
				shell.Advertised(shell.Command(fmt.Sprintf("cd %s", vcsConfig.RepositoryName))),
				shell.Advertised(shell.Command(fmt.Sprintf("hg pull %s", repositoryUrl))),
			),
			shell.Chain(
				shell.Silent(shell.Command(fmt.Sprintf("rm -rf %s", vcsConfig.RepositoryName))),
				shell.Advertised(shell.Command(fmt.Sprintf("hg clone --uncompressed %s %s", repositoryUrl, vcsConfig.RepositoryName))),
				shell.Command(fmt.Sprintf("cd %s", vcsConfig.RepositoryName)),
			),
		),
		shell.Or(
			shell.Advertised(shell.Command(fmt.Sprintf("hg update --clean %s", ref))),
			shell.And(
				shell.Command("mkdir -p .hg/strip-backup"),
				shell.Command(fmt.Sprintf("echo Downloading bundle file for %s", ref)),
				shell.Pipe(
					shell.Command(fmt.Sprintf("ssh -q %s %s", vcsConfig.RepositoryHost, shell.Quote(fmt.Sprintf("hg cat-bundle %s %s", repositoryUrl, ref)))),
					shell.Redirect(
						shell.Command("base64 -d"),
						shell.Command(fmt.Sprintf(".hg/strip-backup/%s", ref)),
						false,
					),
				),
				shell.Advertised(shell.Command(fmt.Sprintf("hg unbundle .hg/strip-backup/%s.hg", ref))),
				shell.Advertised(shell.Command(fmt.Sprintf("hg update --clean %s", ref))),
			),
		),
	)
	return checkoutCommand
}
