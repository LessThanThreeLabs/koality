package events

type Event struct {
	Resource string        `code:"resource"`
	Name     string        `codec:"name"`
	Args     []interface{} `codec:"args"`
}

func NewEvent(resource, name string, args ...interface{}) *Event {
	return &Event{resource, name, args}
}

type EventError struct {
	Message string
}

func (err EventError) Error() string {
	return err.Message
}
