package main

import (
	"fmt"
	"koality/build/debuginstancerunner"
	"koality/build/testrunner"
	"koality/github"
	"koality/internalapi"
	"koality/mail"
	"koality/repositorymanager"
	"koality/resources"
	"koality/resources/database"
	"koality/resources/database/migrate"
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
	if err = poolManager.SubscribeToEvents(resourcesConnection, ec2Broker); err != nil {
		panic(err)
	}

	repositoryManager := repositorymanager.New(repositoriesPath, resourcesConnection)

	testRunner := testrunner.New(resourcesConnection, poolManager, repositoryManager)
	if err = testRunner.SubscribeToEvents(); err != nil {
		panic(err)
	}

	debugInstanceRunner := debuginstancerunner.New(resourcesConnection, poolManager, repositoryManager)
	if err = debugInstanceRunner.SubscribeToEvents(); err != nil {
		panic(err)
	}

	// TODO: initialize more components here
	internalapi.Start(resourcesConnection, poolManager, repositoriesPath, internalapi.RpcSocket)

	gitHubOAuthConnection := github.NewCompoundGitHubOAuthConnection(resourcesConnection)
	gitHubConnection := github.NewConnection(resourcesConnection, gitHubOAuthConnection)
	if err = gitHubConnection.SubscribeToEvents(); err != nil {
		panic(err)
	}

	webserver, err := webserver.New(resourcesConnection, repositoryManager, gitHubConnection, webserverPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("Koality successfully started!")

	// This will block
	if err = webserver.Start(); err != nil {
		panic(err)
	}
}

func setupResourcesConnection(resourcesConnection *resources.Connection) {

}
