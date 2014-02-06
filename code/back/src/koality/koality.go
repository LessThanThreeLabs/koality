package main

import (
	"flag"
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
	"koality/vm/ec2/ec2vm"
	"koality/vm/localmachine"
	"koality/vm/poolmanager"
	"koality/webserver"
	"runtime"
)

const (
	webserverPort = 8080
	koalityRoot   = "/etc/koality"
)

var useEc2Flag = flag.Bool("ec2", false, "Use EC2 pool instead of the fake vm pool")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

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

	ec2Broker := ec2broker.New()

	// TODO (bbland): use the real pool configuration
	var virtualMachinePools []vm.VirtualMachinePool
	if *useEc2Flag {
		ec2Pools, err := resourcesConnection.Pools.Read.GetAllEc2Pools()
		if err != nil {
			panic(err)
		} else if len(ec2Pools) == 0 {
			panic("No ec2 pools configured")
		}

		virtualMachinePools = make([]vm.VirtualMachinePool, len(ec2Pools))
		for index, ec2Pool := range ec2Pools {
			ec2Manager, err := ec2vm.NewManager(ec2Broker, &ec2Pool, resourcesConnection)
			if err != nil {
				panic(err)
			}

			virtualMachinePools[index] = ec2vm.NewPool(ec2Manager)
		}
	} else {
		virtualMachinePools = []vm.VirtualMachinePool{vm.NewPool(0, localmachine.Manager, 0, 3)}
	}

	poolManager := poolmanager.New(virtualMachinePools)

	if err = poolManager.SubscribeToEvents(resourcesConnection, ec2Broker); err != nil {
		panic(err)
	}

	repositoryManager := repositorymanager.New(koalityRoot, resourcesConnection)

	testRunner := testrunner.New(resourcesConnection, poolManager, repositoryManager)
	if err = testRunner.SubscribeToEvents(); err != nil {
		panic(err)
	}

	debugInstanceRunner := debuginstancerunner.New(resourcesConnection, poolManager, repositoryManager)
	if err = debugInstanceRunner.SubscribeToEvents(); err != nil {
		panic(err)
	}

	// TODO: initialize more components here
	internalapi.Start(resourcesConnection, poolManager, koalityRoot, internalapi.RpcSocket)

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
