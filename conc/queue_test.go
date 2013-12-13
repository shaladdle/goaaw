package conc

import (
    "testing"
)

func TestQueue(t *testing.T) {
    q := NewQueue()
    const n = 100
    for i := 0; i < n; i++ {
        t.Log(i)
        q.Push(i)
    }
    for i := 0; i < n; i++ {
        if got := q.Pop().(int); got != i {
            t.Errorf("got %d, wanted %d", got, i)
        }
    }
}
