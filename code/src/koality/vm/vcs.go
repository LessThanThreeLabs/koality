package vm

import (
	"fmt"
	"koality/shell"
	"os/exec"
)

// TODO (bbland): move this somewhere else
type Executor interface {
	Execute(shell.Command) exec.Cmd
}

type VcsConfig struct {
	RepositoryName string
	RepositoryHost string
	RepositoryPath string
}

type Vcs interface {
	Clone(Executor) exec.Cmd
	Checkout(Executor, ref string) exec.Cmd
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
			),
		),
	)
}

type Git struct {
	VcsConfig
}

func (git Git) Clone(executor Executor) exec.Cmd {
	repositoryUrl := fmt.Sprintf("%s:%s", git.VcsConfig.RepositoryHost, git.VcsConfig.RepositoryPath)
	cloneCommand := shell.And(
		shell.Or(
			shell.Not(shell.Test(shell.Command(fmt.Sprintf("-e %s", git.VcsConfig.RepositoryName)))),
			shell.Advertised(shell.Command(fmt.Sprintf("rm -rf %s", git.VcsConfig.RepositoryName))),
		),
		hostAccessCheckCommand(git.VcsConfig.RepositoryHost),
		shell.Advertised(shell.Command(fmt.Sprintf("git clone %s %s", repositoryUrl, git.VcsConfig.RepositoryName))),
	)
	return executor.Execute(cloneCommand)
}

func (git Git) Checkout(executor Executor, ref string) exec.Cmd {
	repositoryUrl := fmt.Sprintf("%s:%s", git.VcsConfig.RepositoryHost, git.VcsConfig.RepositoryPath)
	checkoutCommand := shell.And(
		hostAccessCheckCommand(git.VcsConfig.RepositoryHost),
		shell.Or(
			shell.Silent(shell.Command("which git")),
			shell.Sudo(shell.Command("apt-get install -y git")), // TODO: make this not just apt
		),
		shell.Or(
			shell.Silent(shell.Command(fmt.Sprintf("mv /repositories/cached/%s %s", git.VcsConfig.RepositoryName, git.VcsConfig.RepositoryName))),
			shell.Chain(
				shell.Silent(shell.Command(fmt.Sprintf("rm -rf %s", git.VcsConfig.RepositoryName))),
				shell.Advertised(shell.Command(fmt.Sprintf("git init %s", git.VcsConfig.RepositoryName))),
			),
		),
		shell.Advertised(shell.Command(fmt.Sprintf("cd %s", git.VcsConfig.RepositoryName))),
		shell.Advertised(shell.Command(fmt.Sprintf("git fetch %s %s -n --depth 1", repositoryUrl, ref))),
		shell.Advertised(shell.Command("git checkout --force FETCH_HEAD")),
	)
	return executor.Execute(checkoutCommand)
}

type Hg struct {
	VcsConfig
}

func (hg Hg) Clone(executor Executor) exec.Cmd {
	repositoryUrl := fmt.Sprintf("%s/%s", hg.VcsConfig.RepositoryHost, hg.VcsConfig.RepositoryPath)
	cloneCommand := shell.And(
		shell.Or(
			shell.Not(shell.Test(shell.Command(fmt.Sprintf("-e %s", hg.VcsConfig.RepositoryName)))),
			shell.Advertised(shell.Command(fmt.Sprintf("rm -rf %s", hg.VcsConfig.RepositoryName))),
		),
		hostAccessCheckCommand(hg.VcsConfig.RepositoryHost),
		shell.Advertised(shell.Command(fmt.Sprintf("hg clone --uncompressed %s %s", repositoryUrl, hg.VcsConfig.RepositoryName))),
	)
	return executor.Execute(cloneCommand)
}

func (hg Hg) CheckoutCommand(executor Executor, ref string) exec.Cmd {
	repositoryUrl := fmt.Sprintf("%s/%s", hg.VcsConfig.RepositoryHost, hg.VcsConfig.RepositoryPath)
	checkoutCommand := shell.And(
		hostAccessCheckCommand(hg.VcsConfig.RepositoryHost),
		shell.Or(
			shell.Silent(shell.Command("which hg")),
			shell.Sudo(shell.Command("apt-get install -y mercurial")), // TODO: make this not just apt
		),
		shell.IfElse(
			shell.Silent(shell.Command(fmt.Sprintf("mv /repositories/cached/%s %s", hg.VcsConfig.RepositoryName, hg.VcsConfig.RepositoryName))),
			shell.And(
				shell.Advertised(shell.Command(fmt.Sprintf("cd %s", hg.VcsConfig.RepositoryName))),
				shell.Advertised(shell.Command(fmt.Sprintf("hg pull %s", repositoryUrl))),
			),
			shell.Chain(
				shell.Silent(shell.Command(fmt.Sprintf("rm -rf %s", hg.VcsConfig.RepositoryName))),
				shell.Advertised(shell.Command(fmt.Sprintf("hg clone --uncompressed %s %s", repositoryUrl, hg.VcsConfig.RepositoryName))),
				shell.Command(fmt.Sprintf("cd %s", hg.VcsConfig.RepositoryName)),
			),
		),
		shell.Or(
			shell.Advertised(shell.Command(fmt.Sprintf("hg update --clean %s", ref))),
			shell.And(
				shell.Command("mkdir -p .hg/strip-backup"),
				shell.Command(fmt.Sprintf("echo Downloading bundle file for %s", ref)),
				shell.Pipe(
					shell.Command(fmt.Sprintf("ssh -q %s %s", hg.VcsConfig.RepositoryHost, shell.Quote(fmt.Sprintf("hg cat-bundle %s %s", repositoryUrl, ref)))),
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
	return executor.Execute(checkoutCommand)
}
