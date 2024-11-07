package domain

type Event interface {
	Key() string
}

type EventCallback func(Event)
