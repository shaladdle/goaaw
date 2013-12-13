package conc

import (
    "container/list"

    "github.com/shaladdle/goaaw/container"
)

func NewQueue() container.Queue {
    q :=  &queue{
        push: make(chan interface{}),
        pop: make(chan chan interface{}),
    }

    go q.director()

    return q
}

type queue struct {
    push chan interface{}
    pop chan chan interface{}
}

func (q *queue) director() {
    lst := list.New()
    pops := list.New()

    for {
        select {
        case value := <-q.push:
            if pops.Len() == 0 {
                lst.PushBack(value)
            } else {
                reply := pops.Remove(pops.Front()).(chan interface{})
                reply <- value
            }
        case reply := <-q.pop:
            if lst.Len() > 0 {
                reply <- lst.Remove(lst.Front())
            } else {
                pops.PushBack(reply)
            }
        }
    }
}

func (q *queue) Push(value interface{}) {
    q.push <- value
}

func (q *queue) Pop() interface{} {
    reply := make(chan interface{})
    q.pop <- reply
    return <-reply
}
