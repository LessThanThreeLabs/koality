package main

import (
	"fmt"
	"koality/internal_api"
	"koality/resources/database"
	verificationrunner "koality/verification/runner"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"koality/vm/localmachine"
	"koality/webserver"
	"runtime"
)

const (
	webserverPort = 8080
	rpcSocket     = "/tmp/koality-rpc.sock"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

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
	internal_api.Setup(resourcesConnection, rpcSocket)

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
