package hub

import (
	"time"
)

type Event[T interface{}] struct {
	Text string    `json:"text"`
	Data T         `json:"data"`
	Time time.Time `json:"time"`
}

func NewEvent[T interface{}](text string, data T, time time.Time) *Event[T] {
	return &Event[T]{
		Text: text,
		Data: data,
		Time: time,
	}
}

func CreateEvent[T interface{}](text string, data T) *Event[T] {
	return NewEvent[T](text, data, time.Now())
}
