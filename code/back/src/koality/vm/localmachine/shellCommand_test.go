package localmachine

import (
	"fmt"
	"koality/shell"
	"testing"
	"time"
)

func TestSimpleCommands(testing *testing.T) {
	executeAndAssert(testing, shell.Command("true"), true)
	executeAndAssert(testing, shell.Command("false"), false)

	executeAndAssert(testing, shell.Command("echo hello world"), true)
	executeAndAssert(testing, shell.Command("cat ."), false)
}

func TestAndCommand(testing *testing.T) {
	executeAndAssert(testing,
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
	executeAndAssert(testing, shell.And(trueList...), true)

	failAllList := make([]shell.Command, 20)
	for x := 0; x < len(failAllList); x++ {
		failAllList[x] = shell.Command("false")
	}
	executeAndAssert(testing, shell.And(failAllList...), false)

	failFirstList := make([]shell.Command, 20)
	failFirstList[0] = shell.Command("false")
	for x := 1; x < len(failFirstList); x++ {
		failFirstList[x] = shell.Command("true")
	}
	executeAndAssert(testing, shell.And(failFirstList...), false)

	failLastList := make([]shell.Command, 20)
	for x := 0; x < len(failLastList)-1; x++ {
		failLastList[x] = shell.Command("true")
	}
	failLastList[len(failLastList)-1] = shell.Command("false")
	executeAndAssert(testing, shell.And(failLastList...), false)

	failMiddleList := make([]shell.Command, 20)
	for x := 0; x < len(failMiddleList); x++ {
		failMiddleList[x] = shell.Command("true")
	}
	failMiddleList[len(failMiddleList)/2] = shell.Command("false")
	executeAndAssert(testing, shell.And(failMiddleList...), false)
}

func TestOrCommand(testing *testing.T) {
	executeAndAssert(testing,
		shell.Or(
			shell.Command("true"),
			shell.Command("false"),
		),
		true,
	)

	executeAndAssert(testing,
		shell.Or(
			shell.Command("false"),
			shell.Command("true"),
		),
		true,
	)

	executeAndAssert(testing,
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
	executeAndAssert(testing, shell.And(passAllList...), true)

	failAllList := make([]shell.Command, 20)
	for x := 0; x < len(failAllList); x++ {
		failAllList[x] = shell.Command("false")
	}
	executeAndAssert(testing, shell.And(failAllList...), false)

	passFirstList := make([]shell.Command, 20)
	passFirstList[0] = shell.Command("true")
	for x := 1; x < len(passFirstList); x++ {
		passFirstList[x] = shell.Command("false")
	}
	executeAndAssert(testing, shell.And(passFirstList...), false)

	passLastList := make([]shell.Command, 20)
	for x := 0; x < len(passLastList)-1; x++ {
		passLastList[x] = shell.Command("false")
	}
	passLastList[len(passLastList)-1] = shell.Command("true")
	executeAndAssert(testing, shell.And(passLastList...), false)

	passMiddleList := make([]shell.Command, 20)
	for x := 0; x < len(passMiddleList); x++ {
		passMiddleList[x] = shell.Command("false")
	}
	passMiddleList[len(passMiddleList)/2] = shell.Command("true")
	executeAndAssert(testing, shell.And(passMiddleList...), false)
}

func TestNotCommand(testing *testing.T) {
	executeAndAssert(testing, shell.Not(shell.Command("false")), true)
	executeAndAssert(testing, shell.Not(shell.Command("true")), false)

	executeAndAssert(testing, shell.Not(shell.Command("echo hi")), false)

	executeAndAssert(testing,
		shell.Not(
			shell.And(
				shell.Command("true"),
				shell.Command("false"),
			),
		),
		true,
	)

	executeAndAssert(testing,
		shell.Not(
			shell.And(
				shell.Command("false"),
				shell.Command("true"),
			),
		),
		true,
	)

	executeAndAssert(testing,
		shell.Not(
			shell.And(
				shell.Command("false"),
				shell.Command("false"),
			),
		),
		true,
	)

	executeAndAssert(testing,
		shell.Not(
			shell.And(
				shell.Command("true"),
				shell.Command("true"),
			),
		),
		false,
	)
}

func TestChainCommand(testing *testing.T) {
	executeAndAssert(testing, shell.Chain(shell.Command("true")), true)
	executeAndAssert(testing, shell.Chain(shell.Command("false")), false)

	executeAndAssert(testing,
		shell.Chain(
			shell.Command("false"),
			shell.Command("true"),
		), true)
	executeAndAssert(testing,
		shell.Chain(
			shell.Command("true"),
			shell.Command("false"),
		), false)

	passFirstList := make([]shell.Command, 20)
	passFirstList[0] = shell.Command("true")
	for x := 1; x < len(passFirstList); x++ {
		passFirstList[x] = shell.Command("false")
	}
	executeAndAssert(testing, shell.Chain(passFirstList...), false)

	passLastList := make([]shell.Command, 20)
	for x := 0; x < len(passLastList)-1; x++ {
		passLastList[x] = shell.Command("false")
	}
	passLastList[len(passLastList)-1] = shell.Command("true")
	executeAndAssert(testing, shell.Chain(passLastList...), true)

	passMiddleList := make([]shell.Command, 20)
	for x := 0; x < len(passMiddleList); x++ {
		passMiddleList[x] = shell.Command("false")
	}
	passMiddleList[len(passMiddleList)/2] = shell.Command("true")
	executeAndAssert(testing, shell.Chain(passMiddleList...), false)
}

func TestBackgroundCommand(testing *testing.T) {
	executeWithTimeout(testing, shell.Background(shell.Command("sleep 10")), true, time.Second)
	executeWithTimeout(testing, shell.Background(shell.Command("false")), true, time.Second)

	executeWithTimeout(testing,
		shell.Chain(
			shell.Background("sleep 10"),
			shell.Command("true"),
		),
		true, time.Second)
	executeWithTimeout(testing,
		shell.Chain(
			shell.Background("sleep 10"),
			shell.Command("false"),
		),
		false, time.Second)

	executeWithTimeout(testing,
		shell.Background(
			shell.And(
				shell.Command("true"),
				shell.Command("sleep 10"),
				shell.Command("true"),
			),
		),
		true, time.Second)
}

func TestCaptureCommand(testing *testing.T) {
	executeAndAssert(testing, shell.Capture(shell.Command("echo true")), true)
	executeAndAssert(testing, shell.Capture(shell.Command("echo false")), false)

	executeAndAssert(testing,
		shell.Capture(
			shell.Or(
				shell.Command("false"),
				shell.Command("echo true"),
			),
		), true)
	executeAndAssert(testing,
		shell.Capture(
			shell.Or(
				shell.Command("false"),
				shell.Command("echo false"),
			),
		), false)
}

func TestTestCommand(testing *testing.T) {
	executeAndAssert(testing, shell.Test(shell.Command("-d .")), true)
	executeAndAssert(testing, shell.Test(shell.Command("-f .")), false)

	executeAndAssert(testing, shell.Test(shell.Command("a == a")), true)
	executeAndAssert(testing, shell.Test(shell.Command("a != a")), false)

	executeAndAssert(testing, shell.Not(shell.Test(shell.Command("a != a"))), true)
	executeAndAssert(testing, shell.Not(shell.Test(shell.Command("a == a"))), false)
}

func TestIfCommand(testing *testing.T) {
	executeWithTimeout(testing,
		shell.If(
			shell.Command("false"),
			shell.Command("sleep 10"),
		), true, time.Second)

	executeWithTimeout(testing,
		shell.If(
			shell.Command("true"),
			shell.Command("true"),
		), true, time.Second)
	executeWithTimeout(testing,
		shell.If(
			shell.Command("true"),
			shell.Command("false"),
		), false, time.Second)
}

func TestIfElseCommand(testing *testing.T) {
	executeWithTimeout(testing,
		shell.IfElse(
			shell.Command("true"),
			shell.Command("true"),
			shell.Command("sleep 10"),
		), true, time.Second)
	executeWithTimeout(testing,
		shell.IfElse(
			shell.Command("true"),
			shell.Command("false"),
			shell.Command("sleep 10"),
		), false, time.Second)

	executeWithTimeout(testing,
		shell.IfElse(
			shell.Command("false"),
			shell.Command("sleep 10"),
			shell.Command("true"),
		), true, time.Second)
	executeWithTimeout(testing,
		shell.IfElse(
			shell.Command("false"),
			shell.Command("sleep 10"),
			shell.Command("false"),
		), false, time.Second)
}

func TestPipeCommand(testing *testing.T) {
	executeAndAssert(testing,
		shell.Pipe(
			shell.Command("pwd"),
			shell.Command("xargs ls"),
		), true)
	executeAndAssert(testing,
		shell.Pipe(
			shell.Command("pwd"),
			shell.Command("xargs cat"),
		), false)
}

func TestRedirectCommand(testing *testing.T) {
	mktempCmd := shell.Command(
		fmt.Sprintf("temp=%s",
			shell.Capture(shell.Command("TMPDIR=. mktemp")),
		),
	)

	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo hi"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)

	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)

	// Write to file then overwrite with nothing
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Redirect(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)

	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Redirect(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Redirect(shell.Redirect(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
}

func TestAppendCommand(testing *testing.T) {
	mktempCmd := shell.Command(
		fmt.Sprintf("temp=%s",
			shell.Capture(shell.Command("TMPDIR=. mktemp")),
		),
	)

	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo hi"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)

	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo"), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)

	// Write to file then append nothing
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Command("echo hi"), shell.Command("$temp"), true),
			shell.Append(shell.Command("echo"), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)

	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Append(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), true),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), true)
	executeAndAssert(testing,
		shell.And(
			mktempCmd,
			shell.Command("rm $temp"),
			shell.Append(shell.Append(shell.Command("echo hi"), shell.Command("/dev/stderr"), false), shell.Command("$temp"), false),
			shell.Test(shell.Capture(shell.Command("cat $temp"))),
		), false)
}

func TestSilentCommand(testing *testing.T) {
	executeAndAssert(testing, shell.Silent(shell.Command("echo hi")), true)
	executeAndAssert(testing, shell.Silent(shell.Command("false")), false)

	executeAndAssert(testing, shell.Test(shell.Capture(shell.Silent(shell.Command("echo hi")))), false)

	executeAndAssert(testing, shell.Test(shell.Capture(
		shell.And(
			shell.Command("echo hi"),
			shell.Silent(shell.Command("echo bye")),
		))), true)
	executeAndAssert(testing, shell.Test(shell.Capture(
		shell.Silent(
			shell.And(
				shell.Command("echo hi"),
				shell.Command("echo bye"),
			),
		))), false)
}

func TestAdvertisedCommand(testing *testing.T) {
	executeAndAssert(testing, shell.Advertised(shell.Command("echo hi there")), true)
	executeAndAssert(testing, shell.Advertised(shell.Command("echo 'hi there'")), true)
	executeAndAssert(testing, shell.Advertised(shell.Command("echo \"hi there\"")), true)
	executeAndAssert(testing, shell.Advertised(shell.Command("false")), false)

	executeAndAssert(testing, shell.Advertised(
		shell.And(
			shell.Command("echo hi"),
			shell.Command("echo bye"),
		)), true)

	executeAndAssert(testing, shell.Advertised(
		shell.And(
			shell.Command("false"),
			shell.Command("this wont run"),
		)), false)

	executeAndAssert(testing, shell.AdvertisedWithActual(shell.Command("false"), shell.Command("true")), true)
	executeAndAssert(testing, shell.AdvertisedWithActual(shell.Command("true"), shell.Command("false")), false)
}

func TestLoginCommand(testing *testing.T) {
	executeAndAssert(testing, shell.Login(shell.Command("echo hi there")), true)
	executeAndAssert(testing, shell.Login(shell.Command("echo 'hi there'")), true)
	executeAndAssert(testing, shell.Login(shell.Command("echo \"hi there\"")), true)
	executeAndAssert(testing, shell.Login(shell.Command("false")), false)

	checkLoginCommand := shell.Command("shopt -q login_shell")

	executeAndAssert(testing, shell.Login(checkLoginCommand), true)
	executeAndAssert(testing, shell.Login(shell.And(checkLoginCommand, checkLoginCommand)), true)
	executeAndAssert(testing, checkLoginCommand, false)
}

func TestSudoCommand(testing *testing.T) {
	// Do nothing :(
}

func executeAndAssert(testing *testing.T, command shell.Command, expectSuccess bool) {
	vm := New()
	defer vm.Terminate()

	var err error
	executable, err := vm.MakeExecutable(command, nil, nil, nil)
	if err != nil {
		testing.Logf("Failed to create executable from command:\n%s\n", command)
		testing.Error(err)
	}
	err = executable.Run()
	if expectSuccess {
		if err != nil {
			testing.Logf("Expected command to pass:\n%s\n", command)
			testing.Error(err)
		}
	} else {
		if err == nil {
			testing.Errorf("Expected command to fail:\n%s\n", command)
		}
	}
}

func executeWithTimeout(testing *testing.T, command shell.Command, expectSuccess bool, timeout time.Duration) {
	timeoutChan := time.After(timeout)
	successChan := make(chan bool)
	doneChan := make(chan bool)

	go func() {
		select {
		case <-timeoutChan:
			testing.Errorf("Command timed out after %v:\n%s\n", timeout, command)
			doneChan <- false
			return
		case <-successChan:
			doneChan <- true
			return
		}
	}()

	go func() {
		executeAndAssert(testing, command, expectSuccess)
		successChan <- true
	}()

	<-doneChan
}
