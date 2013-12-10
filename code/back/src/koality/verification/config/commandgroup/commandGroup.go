package commandgroup

import (
	"koality/verification"
	"sync"
)

type CommandGroup interface {
	Next() (verification.Command, error)
	Done() error
	Wait() error
	HasStarted() bool
	Copy() CommandGroup
	Remaining() CommandGroup
}

type AppendableCommandGroup interface {
	CommandGroup
	// TODO (bbland): rename this because it's strange behavior now?
	Append(verification.Command) (verification.Command, error)
}

type appendableCommandGroup struct {
	commands     []verification.Command
	commandIndex uint
	locker       sync.Locker
	waitGroup    *sync.WaitGroup
	hasStarted   bool
}

func New(commands []verification.Command) *appendableCommandGroup {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(commands))
	return &appendableCommandGroup{
		commands:  dedupeCommands(commands),
		locker:    new(sync.Mutex),
		waitGroup: waitGroup,
	}
}

func (group *appendableCommandGroup) Next() (verification.Command, error) {
	group.locker.Lock()
	defer group.locker.Unlock()

	group.hasStarted = true

	if group.commandIndex >= uint(len(group.commands)) {
		return nil, NoMoreCommands
	}
	group.commandIndex++

	return group.commands[group.commandIndex-1], nil
}

func (group *appendableCommandGroup) Done() error {
	group.waitGroup.Done()
	return nil
}

func (group *appendableCommandGroup) Wait() error {
	for uint(len(group.commands)) > group.commandIndex {
		group.waitGroup.Wait()
	}
	return nil
}

func (group *appendableCommandGroup) HasStarted() bool {
	group.locker.Lock()
	defer group.locker.Unlock()

	return group.hasStarted
}

func (group *appendableCommandGroup) Copy() CommandGroup {
	return New(group.commands)
}

func (group *appendableCommandGroup) Remaining() CommandGroup {
	group.locker.Lock()
	defer group.locker.Unlock()

	group.hasStarted = true

	remainingGroup := &appendableCommandGroup{group.commands, group.commandIndex, group.locker, group.waitGroup, group.hasStarted}

	group.commandIndex = uint(len(group.commands))

	return remainingGroup
}

func (group *appendableCommandGroup) Append(command verification.Command) (verification.Command, error) {
	group.locker.Lock()
	defer group.locker.Unlock()

	group.waitGroup.Add(1)
	group.commands = dedupeCommands(append(group.commands, command))

	return group.commands[len(group.commands)-1], nil
}

type noMoreCommands struct{}

func (err noMoreCommands) Error() string {
	return "No more commands"
}

var NoMoreCommands noMoreCommands
