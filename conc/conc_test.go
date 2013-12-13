package conc

import (
    "testing"
    "fmt"
)

func producer(link ProducerChan) {
    for i := 0; i < 100; i++ {
        link.Push(i)
    }
    link.NotifyDone()
}

func consumer(link ConsumerChan, errc chan error) {
    i := 0
    for {
        tag, data := link.Pop()
        switch tag {
        case Value:
            if got := data.(int); got != i {
                errc <- fmt.Errorf("got %d, want %d", got, i)
            }
            i++
        case Error:
            errc <- data.(error)
            break
        case Close:
            fmt.Println("close")
            close(errc)
            break
        }
    }
}

func TestProducerConsumerLink(t *testing.T) {
    p, c := NewStream()
    errc := make(chan error)

    go producer(p)
    go consumer(c, errc)

    if err, ok := <-errc; ok {
        t.Errorf("consumer error: %v", err)
    }
}
