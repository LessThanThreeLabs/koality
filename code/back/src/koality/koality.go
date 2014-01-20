package main

import (
	"fmt"
	"koality/resources/database"
	verificationrunner "koality/verification/runner"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"koality/vm/localmachine"
	"koality/webserver"
)

const (
	webserverPort = 8080
)

func main() {
	resourcesConnection, err := database.New()
	if err != nil {
		panic(err)
	} else if err = database.Migrate(); err != nil {
		panic(err)
	}
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

	fmt.Println("Koality successfully started!")

	// This will block
	err = webserver.Start(resourcesConnection, webserverPort)
	if err != nil {
		panic(err)
	}
}
