package local

import (
    "fmt"
    "reflect"
    "testing"
)

func TestSizeLimit(t *testing.T) {
    maxSize := int64(10)

    type elem struct {
        name string
        size int64
    }

    a := elem{"a", 2}
    b := elem{"b", 7}
    c := elem{"c", 3}
    d := elem{"d", 3}

    order := []elem{a, c, d, d, a, c, a, b}
    wantEvict := []elem{c, d}

    cache := newCache(maxSize)

    for _, el := range order {
        fmt.Println(cache.Get(el.name, el.size))
    }

    if actEvict := cache.Pop(); !reflect.DeepEqual(wantEvict, actEvict) {
        t.Fatalf("Wanted %v, got %v'", wantEvict, actEvict)
    }
}
