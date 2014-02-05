package main

import (
	"fmt"
	"koality/github"
	"koality/internalapi"
	"koality/mail"
	"koality/repositorymanager"
	"koality/resources"
	"koality/resources/database"
	"koality/resources/database/migrate"
	verificationrunner "koality/verification/runner"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"koality/vm/localmachine"
	"koality/vm/poolmanager"
	"koality/webserver"
	"runtime"
)

const (
	webserverPort    = 8080
	repositoriesPath = "/etc/koality"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	resourcesConnection, err := database.New()
	if err != nil {
		panic(err)
	}
	defer resourcesConnection.Close()

	if err := database.Migrate(migrate.Migrations); err != nil {
		panic(err)
	}

	database.KeepClean(resourcesConnection)

	smtpServerSettings, err := resourcesConnection.Settings.Read.GetSmtpServerSettings()
	// Ignore error for now
	mailer := mail.NewMailer(smtpServerSettings)
	mailer.SubscribeToEvents(resourcesConnection)

	// TODO (bbland): use a real pool instead of this bogus one (although this is nice and fast/free)
	virtualMachinePool := vm.NewPool(0, localmachine.Manager, 0, 3)
	poolManager := poolmanager.New([]vm.VirtualMachinePool{virtualMachinePool})

	ec2Broker := ec2broker.New()
	err = poolManager.SubscribeToEvents(resourcesConnection, ec2Broker)
	if err != nil {
		panic(err)
	}

	repositoryManager := repositorymanager.New(repositoriesPath, resourcesConnection)

	verificationRunner := verificationrunner.New(resourcesConnection, poolManager, repositoryManager)
	err = verificationRunner.SubscribeToEvents()
	if err != nil {
		panic(err)
	}

	// TODO: initialize more components here
	internalapi.Start(resourcesConnection, poolManager, repositoriesPath, internalapi.RpcSocket)

	gitHubOAuthConnection := github.NewCompoundGitHubOAuthConnection(resourcesConnection)
	gitHubConnection := github.NewConnection(resourcesConnection, gitHubOAuthConnection)

	webserver, err := webserver.New(resourcesConnection, repositoryManager, gitHubConnection, webserverPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("Koality successfully started!")

	// This will block
	err = webserver.Start()
	if err != nil {
		panic(err)
	}
}

func setupResourcesConnection(resourcesConnection *resources.Connection) {

}
