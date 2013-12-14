package conc

import (
    "fmt"

    "github.com/shaladdle/goaaw/container"
)

type StreamTag byte

const (
    Error = StreamTag(iota)
    Close
    Value
    invalidTag
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

type Handler func(interface{}) error

// HandleValue handles Error and Close cases from 'in' by just propagating them
// to 'out'. For the Value case, it calls valueHandler with the popped value.
func Propagate(in ConsumerChan, out ProducerChan, valueHandler Handler) {
    defer fmt.Println("propagate finished")
    for {
        tag, data := in.Pop()
        switch tag {
        case Value:
            if err := valueHandler(data); err != nil {
                out.NotifyError(err)
                return
            }
        case Error:
            out.NotifyError(data.(error))
            return
        case Close:
            out.NotifyDone()
            return
        }
    }
}

func Finish(in ConsumerChan, valueHandler Handler) error {
    defer fmt.Println("finish finished")
    for {
        tag, data := in.Pop()
        switch tag {
        case Value:
            if err := valueHandler(data); err != nil {
                return err
            }
        case Error:
            return data.(error)
        case Close:
            return nil
        }
    }
}

type multiConsumerChan struct {
    in ConsumerChan
    fanout int
    pop chan chan streamElem
}

func newMultiConsumerChan(in ConsumerChan, fanout int) *multiConsumerChan {
    mcc := &multiConsumerChan{
        in: in,
        fanout: fanout,
        pop: make(chan chan streamElem),
    }

    go mcc.director()

    return mcc
}

func (mcc *multiConsumerChan) director() {
    defer fmt.Println("mcc done")
    for {
        // For every pop request from one of our consumers, we'll pop off the
        // root consumer, and send that as a reply. If the thing we popped was
        // an Error or Close.
        reply := <-mcc.pop
        tag, data := mcc.in.Pop()
        se := streamElem{tag, data}
        reply <- se

        // If we get a Close or Error, just wait until all consumers depending
        // on me have been notified, and then return.
        if tag == Error || tag == Close {
            fmt.Println("notified one thread:", tag)
            for i := mcc.fanout - 1; i > 0; i-- {
                (<-mcc.pop) <- se
                fmt.Println("notified one thread:", tag)
            }
            return
        }
    }
}

func (mcc *multiConsumerChan) Pop() (StreamTag, interface{}) {
    reply := make(chan streamElem)
    mcc.pop <- reply
    se := <-reply
    return se.tag, se.data
}

func MultiConsumerChan(in ConsumerChan, fanout int) []ConsumerChan {
    outChans := make([]ConsumerChan, fanout)
    mcc := newMultiConsumerChan(in, fanout)

    for i := range outChans {
        outChans[i] = mcc
    }

    return outChans
}
