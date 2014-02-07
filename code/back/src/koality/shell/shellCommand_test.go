package shell_test

import (
	"fmt"
	"koality/shell"
	"koality/vm/localmachine"
	"testing"
	"time"
)

func TestSimpleCommands(test *testing.T) {
	executeAndAssert(test, shell.Command("true"), true)
	executeAndAssert(test, shell.Command("false"), false)

	executeAndAssert(test, shell.Command("echo hello world"), true)
	executeAndAssert(test, shell.Command("cat ."), false)
}

func TestAndCommand(test *testing.T) {
	executeAndAssert(test,
		shell.And(
			shell.Command("true"),
			shell.Command("ls"),
		),
		true,
	)
	trueList := make([]shell.Command, 20)
	for x := 0; x < len(trueList); x++ {
		trueList[x] = shell.Command("true")
	}
	executeAndAssert(test, shell.And(trueList...), true)

	failAllList := make([]shell.Command, 20)
	for x := 0; x < len(failAllList); x++ {
		failAllList[x] = shell.Command("false")
	}
	executeAndAssert(test, shell.And(failAllList...), false)

	failFirstList := make([]shell.Command, 20)
	failFirstList[0] = shell.Command("false")
	for x := 1; x < len(failFirstList); x++ {
		failFirstList[x] = shell.Command("true")
	}
	executeAndAssert(test, shell.And(failFirstList...), false)

	failLastList := make([]shell.Command, 20)
	for x := 0; x < len(failLastList)-1; x++ {
		failLastList[x] = shell.Command("true")
	}
	failLastList[len(failLastList)-1] = shell.Command("false")
	executeAndAssert(test, shell.And(failLastList...), false)

	failMiddleList := make([]shell.Command, 20)
	for x := 0; x < len(failMiddleList); x++ {
		failMiddleList[x] = shell.Command("true")
	}
	failMiddleList[len(failMiddleList)/2] = shell.Command("false")
	executeAndAssert(test, shell.And(failMiddleList...), false)
}

func TestOrCommand(test *testing.T) {
	executeAndAssert(test,
		shell.Or(
			shell.Command("true"),
			shell.Command("false"),
		),
		true,
	)

	executeAndAssert(test,
		shell.Or(
			shell.Command("false"),
			shell.Command("true"),
		),
		true,
	)

	executeAndAssert(test,
		shell.Or(
			shell.Command("false"),
			shell.Command("false"),
		),
		false,
	)
	passAllList := make([]shell.Command, 20)
	for x := 0; x < len(passAllList); x++ {
		passAllList[x] = shell.Command("true")
	}
	executeAndAssert(test, shell.And(passAllList...), true)

	failAllList := make([]shell.Command, 20)
	for x := 0; x < len(failAllList); x++ {
		failAllList[x] = shell.Command("false")
	}
	executeAndAssert(test, shell.And(failAllList...), false)

	passFirstList := make([]shell.Command, 20)
	passFirstList[0] = shell.Command("true")
	for x := 1; x < len(passFirstList); x++ {
		passFirstList[x] = shell.Command("false")
	}
	executeAndAssert(test, shell.And(passFirstList...), false)

	passLastList := make([]shell.Command, 20)
	for x := 0; x < len(passLastList)-1; x++ {
		passLastList[x] = shell.Command("false")
	}
	passLastList[len(passLastList)-1] = shell.Command("true")
	executeAndAssert(test, shell.And(passLastList...), false)

	passMiddleList := make([]shell.Command, 20)
	for x := 0; x < len(passMiddleList); x++ {
		passMiddleList[x] = shell.Command("false")
	}
	passMiddleList[len(passMiddleList)/2] = shell.Command("true")
	executeAndAssert(test, shell.And(passMiddleList...), false)
}

func TestNotCommand(test *testing.T) {
	executeAndAssert(test, shell.Not(shell.Command("false")), true)
	executeAndAssert(test, shell.Not(shell.Command("true")), false)

	executeAndAssert(test, shell.Not(shell.Command("echo hi")), false)

	executeAndAssert(test,
		shell.Not(
			shell.And(
				shell.Command("true"),
				shell.Command("false"),
			),
		),
		true,
	)

	executeAndAssert(test,
		shell.Not(
			shell.And(
				shell.Command("false"),
				shell.Command("true"),
			),
		),
		true,
	)

	executeAndAssert(test,
		shell.Not(
			shell.And(
				shell.Command("false"),
				shell.Command("false"),
			),
		),
		true,
	)

	executeAndAssert(test,
		shell.Not(
			shell.And(
				shell.Command("true"),
				shell.Command("true"),
			),
		),
		false,
	)
}

func TestChainCommand(test *testing.T) {
	executeAndAssert(test, shell.Chain(shell.Command("true")), true)
	executeAndAssert(test, shell.Chain(shell.Command("false")), false)

	executeAndAssert(test,
		shell.Chain(
			shell.Command("false"),
			shell.Command("true"),
		), true)
	executeAndAssert(test,
		shell.Chain(
			shell.Command("true"),
			shell.Command("false"),
		), false)

	passFirstList := make([]shell.Command, 20)
	passFirstList[0] = shell.Command("true")
	for x := 1; x < len(passFirstList); x++ {
		passFirstList[x] = shell.Command("false")
	}
	executeAndAssert(test, shell.Chain(passFirstList...), false)

	passLastList := make([]shell.Command, 20)
	for x := 0; x < len(passLastList)-1; x++ {
		passLastList[x] = shell.Command("false")
	}
	passLastList[len(passLastList)-1] = shell.Command("true")
	executeAndAssert(test, shell.Chain(passLastList...), true)

	passMiddleList := make([]shell.Command, 20)
	for x := 0; x < len(passMiddleList); x++ {
		passMiddleList[x] = shell.Command("false")
	}
	passMiddleList[len(passMiddleList)/2] = shell.Command("true")
	executeAndAssert(test, shell.Chain(passMiddleList...), false)
}

func TestBackgroundCommand(test *testing.T) {
	executeWithTimeout(test, shell.Background(shell.Command("sleep 10")), true, time.Second)
	executeWithTimeout(test, shell.Background(shell.Command("false")), true, time.Second)

	executeWithTimeout(test,
		shell.Chain(
			shell.Background("sleep 10"),
			shell.Command("true"),
		),
		true, time.Second)
	executeWithTimeout(test,
		shell.Chain(
			shell.Background("sleep 10"),
			shell.Command("false"),
		),
		false, time.Second)

	executeWithTimeout(test,
		shell.Background(
			shell.And(
				shell.Command("true"),
				shell.Command("sleep 10"),
				shell.Command("true"),
			),
		),
		true, time.Second)
}

func TestCaptureCommand(test *testing.T) {
	executeAndAssert(test, shell.Capture(shell.Command("echo true")), true)
	executeAndAssert(test, shell.Capture(shell.Command("echo false")), false)

	executeAndAssert(test,
		shell.Capture(
			shell.Or(
				shell.Command("false"),
				shell.Command("echo true"),
			),
		), true)
	executeAndAssert(test,
		shell.Capture(
			shell.Or(
				shell.Command("false"),
				shell.Command("echo false"),
			),
		), false)
}

func TestTestCommand(test *testing.T) {
	executeAndAssert(test, shell.Test(shell.Command("-d .")), true)
	executeAndAssert(test, shell.Test(shell.Command("-f .")), false)

	executeAndAssert(test, shell.Test(shell.Command("a == a")), true)
	executeAndAssert(test, shell.Test(shell.Command("a != a")), false)

	executeAndAssert(test, shell.Not(shell.Test(shell.Command("a != a"))), true)
	executeAndAssert(test, shell.Not(shell.Test(shell.Command("a == a"))), false)
}

func TestIfCommand(test *testing.T) {
	executeWithTimeout(test,
		shell.If(
			shell.Command("false"),
			shell.Command("sleep 10"),
		), true, time.Second)

	executeWithTimeout(test,
		shell.If(
			shell.Command("true"),
			shell.Command("true"),
		), true, time.Second)
	executeWithTimeout(test,
		shell.If(
			shell.Command("true"),
			shell.Command("false"),
		), false, time.Second)
}

func TestIfElseCommand(test *testing.T) {
	executeWithTimeout(test,
		shell.IfElse(
			shell.Command("true"),
			shell.Command("true"),
			shell.Command("sleep 10"),
		), true, time.Second)
	executeWithTimeout(test,
		shell.IfElse(
			shell.Command("true"),
			shell.Command("false"),
			shell.Command("sleep 10"),
		), false, time.Second)

	executeWithTimeout(test,
		shell.IfElse(
			shell.Command("false"),
			shell.Command("sleep 10"),
			shell.Command("true"),
		), true, time.Second)
	executeWithTimeout(test,
		shell.IfElse(
			shell.Command("false"),
			shell.Command("sleep 10"),
			shell.Command("false"),
		), false, time.Second)
}

func TestPipeCommand(test *testing.T) {
	executeAndAssert(test,
		shell.Pipe(
			shell.Command("pwd"),
			shell.Command("xargs ls"),
		), true)
	executeAndAssert(test,
		shell.Pipe(
			shell.Command("pwd"),
			shell.Command("xargs cat"),
		), false)
}

func TestRedirectCommand(test *testing.T) {
	mktempCmd := shell.Command(
		fmt.Sprintf("temp=%s",
			shell.Capture(shell.Command("TMPDIR=. mktemp")),
		),
	)

	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo hi"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)

	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)

	// Write to file then overwrite with nothing
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Redirect(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)

	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Redirect(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Redirect(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
}

func TestAppendCommand(test *testing.T) {
	mktempCmd := shell.Command(
		fmt.Sprintf("temp=%s",
			shell.Capture(shell.Command("TMPDIR=. mktemp")),
		),
	)

	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo hi"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)

	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)

	// Write to file then append nothing
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Append(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)

	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Append(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(test,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Append(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
}

func TestSilentCommand(test *testing.T) {
	executeAndAssert(test, shell.Silent(shell.Command("echo hi")), true)
	executeAndAssert(test, shell.Silent(shell.Command("false")), false)

	executeAndAssert(test, shell.Test(shell.Capture(shell.Silent(shell.Command("echo hi")))), false)

	executeAndAssert(test, shell.Test(shell.Capture(
		shell.And(
			shell.Command("echo hi"),
			shell.Silent(shell.Command("echo bye")),
		))), true)
	executeAndAssert(test, shell.Test(shell.Capture(
		shell.Silent(
			shell.And(
				shell.Command("echo hi"),
				shell.Command("echo bye"),
			),
		))), false)
}

func TestAdvertisedCommand(test *testing.T) {
	executeAndAssert(test, shell.Advertised(shell.Command("echo hi there")), true)
	executeAndAssert(test, shell.Advertised(shell.Command("echo 'hi there'")), true)
	executeAndAssert(test, shell.Advertised(shell.Command("echo \"hi there\"")), true)
	executeAndAssert(test, shell.Advertised(shell.Command("false")), false)

	executeAndAssert(test, shell.Advertised(
		shell.And(
			shell.Command("echo hi"),
			shell.Command("echo bye"),
		)), true)

	executeAndAssert(test, shell.Advertised(
		shell.And(
			shell.Command("false"),
			shell.Command("this wont run"),
		)), false)

	executeAndAssert(test, shell.AdvertisedWithActual(shell.Command("false"), shell.Command("true")), true)
	executeAndAssert(test, shell.AdvertisedWithActual(shell.Command("true"), shell.Command("false")), false)
}

func TestLoginCommand(test *testing.T) {
	executeAndAssert(test, shell.Login(shell.Command("echo hi there")), true)
	executeAndAssert(test, shell.Login(shell.Command("echo 'hi there'")), true)
	executeAndAssert(test, shell.Login(shell.Command("echo \"hi there\"")), true)
	executeAndAssert(test, shell.Login(shell.Command("false")), false)

	checkLoginCommand := shell.Command("shopt -q login_shell")

	executeAndAssert(test, shell.Login(checkLoginCommand), true)
	executeAndAssert(test, shell.Login(shell.And(checkLoginCommand, checkLoginCommand)), true)
	executeAndAssert(test, checkLoginCommand, false)
}

func TestSudoCommand(test *testing.T) {
	test.Skip("Not really sure how to safely test this...")
}

func executeAndAssert(test *testing.T, command shell.Command, expectSuccess bool) {
	vm, err := localmachine.New()
	if err != nil {
		test.Fatal(err)
	}

	defer vm.Terminate()

	executable := shell.Executable{
		Command: command,
	}
	execution, err := vm.Execute(executable)
	if err != nil {
		test.Logf("Failed to execute command:\n%s\n", command)
		test.Error(err)
	}
	err = execution.Wait()
	if expectSuccess {
		if err != nil {
			test.Logf("Expected command to pass:\n%s\n", command)
			test.Error(err)
		}
	} else {
		if err == nil {
			test.Errorf("Expected command to fail:\n%s\n", command)
		}
	}
}

func executeWithTimeout(test *testing.T, command shell.Command, expectSuccess bool, timeout time.Duration) {
	timeoutChan := time.After(timeout)
	successChan := make(chan bool)
	doneChan := make(chan bool)

	go func() {
		select {
		case <-timeoutChan:
			test.Errorf("Command timed out after %v:\n%s\n", timeout, command)
			doneChan <- false
			return
		case <-successChan:
			doneChan <- true
			return
		}
	}()

	go func() {
		executeAndAssert(test, command, expectSuccess)
		successChan <- true
	}()

	<-doneChan
}
