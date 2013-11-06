package commandgroup

import (
	"koality/verification"
	"sync"
)

type CommandGroup interface {
	HasMoreCommands() bool
	Next() (verification.Command, error)
	Done() error
	Wait() error
}

type AppendableCommandGroup interface {
	CommandGroup
	Append(verification.Command) error
}

type appendableCommandGroup struct {
	commands  []verification.Command
	locker    sync.Locker
	waitGroup *sync.WaitGroup
}

func New(commands []verification.Command) *appendableCommandGroup {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(commands))
	return &appendableCommandGroup{
		commands:  commands,
		locker:    new(sync.Mutex),
		waitGroup: waitGroup,
	}
}

func (group *appendableCommandGroup) HasMoreCommands() bool {
	group.locker.Lock()
	defer group.locker.Unlock()

	return len(group.commands) > 0
}

func (group *appendableCommandGroup) Next() (verification.Command, error) {
	group.locker.Lock()
	defer group.locker.Unlock()

	if len(group.commands) == 0 {
		return nil, NoMoreCommands
	}
	c := group.commands[0]
	group.commands = group.commands[1:]

	return c, nil
}

func (group *appendableCommandGroup) Done() error {
	group.waitGroup.Done()
	return nil
}

func (group *appendableCommandGroup) Wait() error {
	for group.HasMoreCommands() {
		group.waitGroup.Wait()
	}
	return nil
}

func (group *appendableCommandGroup) Append(command verification.Command) error {
	group.locker.Lock()
	defer group.locker.Unlock()

	group.waitGroup.Add(1)
	group.commands = append(group.commands, command)
	return nil
}

type noMoreCommands struct{}

func (err noMoreCommands) Error() string {
	return "No more commands"
}

var NoMoreCommands noMoreCommands
