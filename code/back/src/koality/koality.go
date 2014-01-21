package main

import (
	"fmt"
	"koality/internalapi"
	"koality/resources/database"
	"koality/resources/database/migrate"
	verificationrunner "koality/verification/runner"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"koality/vm/localmachine"
	"koality/webserver"
	"runtime"
)

const (
	webserverPort = 8080
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err = database.Migrate(migrate.Migrations); err != nil {
		panic(err)
	}

	resourcesConnection, err := database.New()
	if err != nil {
		panic(err)
	}
	defer resourcesConnection.Close()
	database.KeepClean(resourcesConnection)

	// TODO (bbland): use a real pool instead of this bogus one (although this is nice and fast/free)
	virtualMachinePool := vm.NewPool(0, localmachine.Launcher, 0, 3)
	ec2Broker := ec2broker.New()
	verificationRunner := verificationrunner.New(resourcesConnection, []vm.VirtualMachinePool{virtualMachinePool}, ec2Broker)
	err = verificationRunner.SubscribeToEvents()
	if err != nil {
		panic(err)
	}

	// TODO: initialize more components here

	internalapi.Setup(resourcesConnection, internalapi.RpcSocket)

	webserver, err := webserver.New(resourcesConnection, webserverPort)
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
