package vcs

import (
	"koality/resources"
	"koality/shell"
	"path"
)

type Hg struct{} // Empty struct?

var ensureHgInstalledCommand = shell.Or(
	shell.Silent("which hg"),
	shell.Advertised(shell.Sudo("apt-get install -y mercurial")),
)

func (hg Hg) CloneCommand(repository *resources.Repository) shell.Command {
	return shell.And(
		ensureHgInstalledCommand,
		shell.Or(
			shell.Not(shell.Test(shell.Commandf("-d %s", repository.Name))),
			shell.Advertised(shell.Commandf("rm -rf %s", repository.Name)),
		),
		shell.Advertised(shell.Commandf("hg clone --uncompressed %s %s", repository.RemoteUri, repository.Name)),
	)
}

func (hg Hg) CheckoutCommand(repository *resources.Repository, ref string) shell.Command {
	return shell.And(
		ensureHgInstalledCommand,
		shell.IfElse(
			shell.Silent(shell.Commandf("mv %s %s", path.Join("/", "koality", "repositories", repository.Name), repository.Name)),
			shell.And(
				shell.Advertised(shell.Commandf("cd %s", repository.Name)),
				shell.Advertised(shell.Commandf("hg pull %s", repository.RemoteUri)),
			),
			shell.Chain(
				shell.Silent(shell.Commandf("rm -rf %s", repository.Name)),
				shell.Advertised(shell.Commandf("hg clone --uncompressed %s %s", repository.RemoteUri, repository.Name)),
				shell.Commandf("cd %s", repository.Name),
			),
		),
		shell.Advertised(shell.Commandf("hg update --clean %s", ref)),
	)
}
