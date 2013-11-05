package main

import (
	// "bytes"
	"fmt"
	// "github.com/crowdmob/goamz/aws"
	// "github.com/crowdmob/goamz/ec2"
	// "io"
	// "io/ioutil"
	"koality/shell"
	// myioutil "koality/util/ioutil"
	// "koality/vm/ec2/ec2broker"
	// "koality/vm/ec2/ec2vm"
	// "koality/vm/vcs"
	// "koality/vm"
	"koality/verification/stageverifier"
	"koality/vm/localmachine"
	"os"
	// "time"
	// "sync"
)

func main() {
	// broker := ec2broker.New(
	// 	ec2.New(
	// 		aws.Auth{
	// 			AccessKey: accessKey,
	// 			SecretKey: secretKey,
	// 		},
	// 		aws.USWest2,
	// 	),
	// )

	// launcher := ec2vm.NewLauncher(broker)

	// pool := vm.NewPool(launcher, 0, 5)

	// ec2Vm := launcher.LaunchVirtualMachine()
	// ec2Vm := pool.Get()

	vm := localmachine.New()

	// instances := broker.Instances()
	// var blogInstance ec2.Instance
	// for _, instance := range instances {
	// 	if instance.Tags[0].Value == "koality blog" {
	// 		blogInstance = instance
	// 		break
	// 	}
	// }

	// ec2vm := ec2vm.New(&blogInstance, broker, "ec2-user")

	// git := vcs.VcsMap[vcs.VcsType("git")]

	// buf := bytes.NewBuffer(make([]byte, 0))

	// cmd := git.CheckoutCommand(vcs.VcsConfig{"spot", "git@github.com", "BrianBland/spot"}, "master"),
	copyExec, _ := vm.FileCopy("execTest.go", "execTest.go")
	cmd := shell.And(
		shell.Advertised(shell.Command("pwd")),
		shell.Advertised(shell.Command("sleep 1")),
		shell.Advertised(shell.Command("ls -la")),
		shell.Advertised(shell.Background(shell.Command("sleep 10"))),
	)
	exec := vm.MakeExecutable(cmd)
	failExec := vm.MakeExecutable(shell.Advertised("cd derp"))

	commandRunner := stageverifier.OutputWritingCommandRunner{os.Stdout}

	success, err := commandRunner.RunCommand(copyExec)
	if success != true || err != nil {
		panic(err)
	}
	success, err = commandRunner.RunCommand(exec)
	if success != true || err != nil {
		panic(err)
	}
	success, err = commandRunner.RunCommand(failExec)
	if success != false || err == nil {
		panic("command was supposed to fail")
	}

	fmt.Println("Ran!")
	vm.Terminate()
}
