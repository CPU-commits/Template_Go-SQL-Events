package bus

import "time"

type EventName string

type Event struct {
	Name     EventName
	Payload  []byte
	Metadata map[string]string
}

type Context struct {
	Data         []byte
	BindData     func(toBind interface{}) error
	Kill         func(reason string) error
	FollowUp     func(delay time.Duration) error
	EventTrigger string
}

type Bus interface {
	Publish(event Event) error
	Subscribe(name EventName, handler func(c Context) error)
	Request(name Event, toBind interface{}) error
}
