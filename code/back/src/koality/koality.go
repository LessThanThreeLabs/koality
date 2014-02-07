package main

import (
	"flag"
	"fmt"
	"koality/build/debuginstancerunner"
	"koality/build/testrunner"
	"koality/github"
	"koality/internalapi"
	"koality/mail"
	"koality/notify"
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

// TODO(bbland): remove this
var useEc2Flag = flag.Bool("ec2", false, "Use EC2 pool instead of the fake vm pool")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	resourcesConnection := getResourcesConnection()

	mailer := getMailer(resourcesConnection)

	ec2Broker := ec2broker.New()

	poolManager := poolmanager.New(getVirtualMachinePools(resourcesConnection, ec2Broker))
	if err := poolManager.SubscribeToEvents(resourcesConnection, ec2Broker); err != nil {
		panic(err)
	}

	repositoryManager := repositorymanager.New(koalityRoot, resourcesConnection)

	testRunner := testrunner.New(resourcesConnection, poolManager, repositoryManager)
	if err := testRunner.SubscribeToEvents(); err != nil {
		panic(err)
	}

	debugInstanceRunner := debuginstancerunner.New(resourcesConnection, poolManager, repositoryManager, notify.New(resourcesConnection, mailer))
	if err := debugInstanceRunner.SubscribeToEvents(); err != nil {
		panic(err)
	}

	internalapi.Start(resourcesConnection, poolManager, koalityRoot, internalapi.RpcSocket)

	fmt.Println("Koality successfully started!")
	startWebserverAndBlock(resourcesConnection, repositoryManager, mailer)
}

func getResourcesConnection() *resources.Connection {
	resourcesConnection, err := database.New()
	if err != nil {
		panic(err)
	}

	if err := database.Migrate(migrate.Migrations); err != nil {
		panic(err)
	}

	database.KeepClean(resourcesConnection)
	return resourcesConnection
}

func getMailer(resourcesConnection *resources.Connection) mail.Mailer {
	smtpServerSettings, err := resourcesConnection.Settings.Read.GetSmtpServerSettings()
	if _, ok := err.(resources.NoSuchSettingError); !ok && err != nil {
		panic(err)
	}

	mailer := mail.NewMailer(smtpServerSettings)
	mailer.SubscribeToEvents(resourcesConnection)
	return mailer
}

func getVirtualMachinePools(resourcesConnection *resources.Connection, ec2Broker *ec2broker.Ec2Broker) []vm.VirtualMachinePool {
	if *useEc2Flag {
		ec2Pools, err := resourcesConnection.Pools.Read.GetAllEc2Pools()
		if err != nil {
			panic(err)
		} else if len(ec2Pools) == 0 {
			panic("No ec2 pools configured")
		}

		virtualMachinePools := make([]vm.VirtualMachinePool, len(ec2Pools))
		for index, ec2Pool := range ec2Pools {
			ec2Manager, err := ec2vm.NewManager(ec2Broker, &ec2Pool, resourcesConnection)
			if err != nil {
				panic(err)
			}

			virtualMachinePools[index] = ec2vm.NewPool(ec2Manager)
		}
		return virtualMachinePools
	} else {
		return []vm.VirtualMachinePool{vm.NewPool(0, localmachine.Manager, 0, 3)}
	}
}

func startWebserverAndBlock(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager, mailer mail.Mailer) {
	gitHubOAuthConnection := github.NewCompoundGitHubOAuthConnection(resourcesConnection)
	gitHubConnection := github.NewConnection(resourcesConnection, gitHubOAuthConnection)
	if err := gitHubConnection.SubscribeToEvents(); err != nil {
		panic(err)
	}

	webserver, err := webserver.New(resourcesConnection, repositoryManager, gitHubConnection, mailer, webserverPort)
	if err != nil {
		panic(err)
	}

	if err = webserver.Start(); err != nil {
		panic(err)
	}
}
