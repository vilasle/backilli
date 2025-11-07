package model

type Event interface {
	Raise() bool
}
