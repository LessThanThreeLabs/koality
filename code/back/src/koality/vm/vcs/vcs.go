package vcs

import (
	"fmt"
	"koality/resources"
	"koality/shell"
)

type Vcs interface {
	CloneCommand(*resources.Repository) shell.Command
	CheckoutCommand(*resources.Repository, string) shell.Command
}

var (
	git = Git{}
	hg  = Hg{}
)

func CloneCommand(repository *resources.Repository) shell.Command {
	switch repository.VcsType {
	case "git":
		return git.CloneCommand(repository)
	case "hg":
		return hg.CloneCommand(repository)
	default:
		panic(fmt.Sprintf("Unexpected repository type: %s", repository.VcsType))
	}
}

func CheckoutCommand(repository *resources.Repository, ref string) shell.Command {
	switch repository.VcsType {
	case "git":
		return git.CheckoutCommand(repository, ref)
	case "hg":
		return hg.CheckoutCommand(repository, ref)
	default:
		panic(fmt.Sprintf("Unexpected repository type: %s", repository.VcsType))
	}
}
