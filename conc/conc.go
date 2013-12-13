package conc

import (
    "github.com/shaladdle/goaaw/container"
)

type StreamTag byte

const (
    Error = StreamTag(iota)
    Close
    Value
)

type ProducerChan interface {
    container.Pusher
    NotifyError(error)
    NotifyDone()
}

type ConsumerChan interface {
    Pop() (StreamTag, interface{})
}

type stream struct {
    q container.Queue
}

type streamElem struct {
    tag StreamTag
    data interface{}
}

func newStream(q container.Queue) (ProducerChan, ConsumerChan) {
    s := &stream{q}
    return s, s
}

func NewStream() (ProducerChan, ConsumerChan) {
    return newStream(NewQueue())
}

// Producer methods
func (s *stream) Push(elem interface{}) {
    s.q.Push(streamElem{tag: Value, data: elem})
}

func (s *stream) NotifyError(err error) {
    s.q.Push(streamElem{tag: Error, data: err})
}

func (s *stream) NotifyDone() {
    s.q.Push(streamElem{tag: Close})
}

// Consumer method
func (s *stream) Pop() (StreamTag, interface{}) {
    elem := s.q.Pop().(streamElem)
    return elem.tag, elem.data
}
