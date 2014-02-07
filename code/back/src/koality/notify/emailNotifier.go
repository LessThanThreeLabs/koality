package notify

import (
	"fmt"
	"koality/resources"
	"koality/shell"
	"koality/vm"
	"os/user"

	"koality/build/runner"
	"koality/mail"
)

type EmailNotifier struct {
	resourcesConnection *resources.Connection
	mailer              mail.Mailer
}

func New(resourcesConnection *resources.Connection, mailer mail.Mailer) Notifier {
	return &EmailNotifier{resourcesConnection, mailer}
}

func (emailNotifier EmailNotifier) Notify(vm vm.VirtualMachine, build *resources.Build, buildData *runner.BuildData) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	domainName, err := emailNotifier.resourcesConnection.Settings.Read.GetDomainName()
	if err != nil {
		return err
	}

	sshString := fmt.Sprintf("ssh %s@%s "+shell.Quote("ssh %d %s"), currentUser.Username, domainName, buildData.BuildConfig.Params.PoolId, vm.Id())
	emailFrom := fmt.Sprintf("koality@%s", domainName)
	err = emailNotifier.mailer.SendMail(emailFrom, []string{"noreply@koalitycode.com"}, []string{build.EmailToNotify}, sshString, sshString)
	return nil
}
