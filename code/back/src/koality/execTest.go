package main

import (
	// "bytes"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	// "io"
	"io/ioutil"
	"koality/shell"
	myioutil "koality/util/ioutil"
	"koality/vm/ec2/ec2broker"
	"koality/vm/ec2/ec2vm"
	// "koality/vm/vcs"
	"koality/vm"
	// "os"
	// "time"
	// "sync"
)

func main() {
	accessKey := "AKIAJMQW32VH2MIGJ3UQ"
	secretKey := "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"
	broker := ec2broker.New(
		ec2.New(
			aws.Auth{
				AccessKey: accessKey,
				SecretKey: secretKey,
			},
			aws.USWest2,
		),
	)

	launcher := ec2vm.NewLauncher(broker)

	pool := vm.NewPool(launcher, 0, 5)

	// ec2Vm := launcher.LaunchVirtualMachine()
	ec2Vm := pool.Get()

	fmt.Printf("%#v\n", ec2Vm)

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
	scpExec := ec2Vm.FileCopy("execTest.go", "execTest.go")
	cmd := shell.And(
		shell.Advertised(shell.Command("ls")),
		shell.Advertised(shell.Command("cat execTest.go")),
	)
	exec := ec2Vm.MakeExecutable(cmd)

	// cmd := ec2vm.Executor().Execute(
	// 	shell.And(
	// 		shell.Command(fmt.Sprintf("cd %s", shell.Capture(shell.Command("mktemp -d")))),
	// 		shell.Command("pwd"),
	// 	),
	// )

	// execMaker := shell.NewShellExecutableMaker()
	// exec := execMaker.MakeExecutable(shell.Command("ls"))

	scpStdout, _ := scpExec.StdoutPipe()
	scpStderr, _ := scpExec.StderrPipe()

	stdout, _ := exec.StdoutPipe()
	stderr, _ := exec.StderrPipe()

	combinedOut := myioutil.CombineReaders(stdout, stderr, scpStdout, scpStderr)

	// go io.Copy(combinedOut, stdout)
	// go io.Copy(combinedOut, stderr)
	// go io.Copy(os.Stdout, outpipe)
	// errpipe, _ := exec.StderrPipe()
	// exec.Stdout = buf
	// exec.Stderr = buf
	scpExec.Run()
	exec.Run()

	// time.Sleep(time.Duration(5) * time.Second)

	bytes, _ := ioutil.ReadAll(combinedOut)
	fmt.Printf("%s", string(bytes))
	// fmt.Printf("%#v", outpipe)
	// io.Copy(os.Stdout, outpipe)

	// exec.Wait()
	// fmt.Printf("bye\n")
}
