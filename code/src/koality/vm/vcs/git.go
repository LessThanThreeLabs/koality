package vcs

import (
	"fmt"
	"koality/shell"
)

type Git struct{} // Empty struct?

func (git Git) Type() VcsType {
	return VcsType("git")
}

func (git Git) CloneCommand(vcsConfig VcsConfig) shell.Command {
	repositoryUrl := fmt.Sprintf("%s:%s", vcsConfig.RepositoryHost, vcsConfig.RepositoryPath)
	cloneCommand := shell.And(
		shell.Or(
			shell.Not(shell.Test(shell.Command(fmt.Sprintf("-e %s", vcsConfig.RepositoryName)))),
			shell.Advertised(shell.Command(fmt.Sprintf("rm -rf %s", vcsConfig.RepositoryName))),
		),
		hostAccessCheckCommand(vcsConfig.RepositoryHost),
		shell.Advertised(shell.Command(fmt.Sprintf("git clone %s %s", repositoryUrl, vcsConfig.RepositoryName))),
	)
	return cloneCommand
}

func (git Git) CheckoutCommand(vcsConfig VcsConfig, ref string) shell.Command {
	repositoryUrl := fmt.Sprintf("%s:%s", vcsConfig.RepositoryHost, vcsConfig.RepositoryPath)
	checkoutCommand := shell.And(
		hostAccessCheckCommand(vcsConfig.RepositoryHost),
		shell.Or(
			shell.Silent(shell.Command("which git")),
			shell.Sudo(shell.Command("apt-get install -y git")), // TODO: make this not just apt
		),
		shell.Or(
			shell.Silent(shell.Command(fmt.Sprintf("mv /repositories/cached/%s %s", vcsConfig.RepositoryName, vcsConfig.RepositoryName))),
			shell.Chain(
				shell.Silent(shell.Command(fmt.Sprintf("rm -rf %s", vcsConfig.RepositoryName))),
				shell.Advertised(shell.Command(fmt.Sprintf("git init %s", vcsConfig.RepositoryName))),
			),
		),
		shell.Advertised(shell.Command(fmt.Sprintf("cd %s", vcsConfig.RepositoryName))),
		shell.Advertised(shell.Command(fmt.Sprintf("git fetch %s %s -n --depth 1", repositoryUrl, ref))),
		shell.Advertised(shell.Command("git checkout --force FETCH_HEAD")),
	)
	return checkoutCommand
}
