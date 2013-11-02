package events

type Event struct {
	Name string        `codec:"name"`
	Args []interface{} `codec:"args"`
}

func NewEvent(name string, args ...interface{}) *Event {
	return &Event{name, args}
}

type EventError struct {
	Message string
}

func (err EventError) Error() string {
	return err.Message
}
