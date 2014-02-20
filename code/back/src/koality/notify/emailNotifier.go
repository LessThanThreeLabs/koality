package notify

import (
	"fmt"
	"html"
	"koality/mail"
	"koality/resources"
	"koality/shell"
	"koality/vm"
	"os/user"
	"strings"
)

const (
	emailFooter = "-The Koality Team\n<a href=\"https://koalitycode.com\">https://koalitycode.com</a>"
)

type EmailNotifier struct {
	resourcesConnection *resources.Connection
	mailer              mail.Mailer
}

func NewEmailNotifier(resourcesConnection *resources.Connection, mailer mail.Mailer) Notifier {
	return &EmailNotifier{resourcesConnection, mailer}
}

func (emailNotifier *EmailNotifier) NotifyBuildStatus(build *resources.Build) error {
	if build.Status != "failed" || build.EmailToNotify == "" {
		return nil
	}

	domainName, err := emailNotifier.resourcesConnection.Settings.Read.GetDomainName()
	if err != nil {
		return err
	}

	branchName := build.Ref
	if strings.HasPrefix(branchName, "refs/heads/") {
		branchName = branchName[len("refs/heads/"):]
	}
	viewBuildUrl := getBuildUri(domainName, build.RepositoryId, build.Id)
	viewBuildLink := "<a href=\"" + viewBuildUrl + "\">" + viewBuildUrl + "</a>"

	emailFrom := fmt.Sprintf("koality@%s", domainName)
	emailSubject := fmt.Sprintf("There was an issue with your change (%s)", build.Changeset.HeadSha)
	commitMessageHtml := "<code>" + html.EscapeString(build.Changeset.HeadMessage) + "</code>"
	emailBody := htmlSanitizer.Replace(
		html.EscapeString(build.Changeset.HeadUsername) + ",\n\n" +
			"There was an issue with your change (" + html.EscapeString(build.Changeset.HeadSha) + ").\n" +
			"Please fix the change and resubmit it.\n\n" +
			"Details for your change are available here: " + viewBuildLink + "\n\n" +
			"Branch: " + html.EscapeString(branchName) + "\n\n" +
			"Commit Message:\n" + commitMessageHtml + "\n\n" +
			emailFooter,
	)

	err = emailNotifier.mailer.SendMail(emailFrom, []string{"noreply@koalitycode.com"}, []string{build.EmailToNotify}, emailSubject, emailBody)
	return err
}

func (emailNotifier *EmailNotifier) NotifyDebugInstance(vm vm.VirtualMachine, build *resources.Build, debugInstance *resources.DebugInstance) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	domainName, err := emailNotifier.resourcesConnection.Settings.Read.GetDomainName()
	if err != nil {
		return err
	}

	sshString := fmt.Sprintf("ssh %s@%s "+shell.Quote("ssh %d %s"), currentUser.Username, domainName, debugInstance.PoolId, vm.Id())
	emailSubject := "Your debug instance is ready"
	emailBody := htmlSanitizer.Replace(
		html.EscapeString(build.Changeset.HeadUsername) + ",\n\n" +
			"Your debug instance has launched and will be accessible until " + html.EscapeString(debugInstance.Expires.String()) + "\n\n" +
			"To SSH into your debug instance, type the following command into your terminal:\n\n" +
			html.EscapeString(sshString) + "\n\n" +
			emailFooter,
	)

	emailFrom := fmt.Sprintf("koality@%s", domainName)
	err = emailNotifier.mailer.SendMail(emailFrom, []string{"noreply@koalitycode.com"}, []string{build.EmailToNotify}, emailSubject, emailBody)
	return err
}
