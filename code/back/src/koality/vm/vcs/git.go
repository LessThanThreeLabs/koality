package vcs

import (
	"koality/resources"
	"koality/shell"
	"path"
)

type Git struct{} // Empty struct?

var ensureGitInstalledCommand = shell.Or(
	shell.Silent("which git"),
	shell.Advertised(shell.Sudo("apt-get install -y git-core")),
)

func (git Git) CloneCommand(repository *resources.Repository) shell.Command {
	return shell.And(
		ensureGitInstalledCommand,
		shell.Or(
			shell.Not(shell.Test(shell.Commandf("-e %s", repository.Name))),
			shell.Advertised(shell.Commandf("rm -rf %s", repository.Name)),
		),
		shell.Advertised(shell.Commandf("git clone %s %s", repository.RemoteUri, repository.Name)),
	)
}

func (git Git) CheckoutCommand(repository *resources.Repository, ref string) shell.Command {
	return shell.And(
		ensureGitInstalledCommand,
		shell.Or(
			shell.And(
				shell.Test(shell.Commandf("-d %s", path.Join("/", "koality", "repositories", repository.Name))),
				shell.Advertised(shell.Commandf("mv %s %s", path.Join("/", "koality", "repositories", repository.Name), repository.Name)),
			),
			shell.Chain(
				shell.Silent(shell.Commandf("rm -rf %s", repository.Name)),
				shell.Advertised(shell.Commandf("git init %s", repository.Name)),
			),
		),
		shell.Advertised(shell.Commandf("cd %s", repository.Name)),
		shell.Advertised(shell.Commandf("git fetch %s %s -n --depth 1", repository.RemoteUri, ref)),
		shell.Advertised("git checkout --force FETCH_HEAD"),
	)
}
