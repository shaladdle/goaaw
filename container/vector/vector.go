package vector

import (
	"container/list"
)

type pair struct {
	key   string
	value interface{}
}

type Vector struct {
	l *list.List
	m map[string]*list.Element
}

func New() *Vector {
	return &Vector{
		l: list.New(),
		m: make(map[string]*list.Element),
	}
}

func (v *Vector) AppendKey(key string) {
	v.m[key] = v.l.PushBack(pair{key, nil})
}

func (v *Vector) Append(key string, val interface{}) {
	v.m[key] = v.l.PushBack(pair{key, val})
}

func (v Vector) Get(key string) interface{} {
	return v.m[key].Value.(pair).value
}

func (v Vector) Has(key string) bool {
	_, has := v.m[key]
	return has
}

func (v Vector) Delete(key string) {
	v.l.Remove(v.m[key])
	delete(v.m, key)
}

func (v Vector) KeySlice() []string {
	ret := make([]string, len(v.m))
	i := 0
	for el := v.l.Front(); el != nil; el = el.Next() {
		ret[i] = el.Value.(pair).key
		i++
	}
	return ret
}

func (v Vector) KeyIter() <-chan string {
	ch := make(chan string)
	go func() {
		for el := v.l.Front(); el != nil; el = el.Next() {
			ch <- el.Value.(pair).key
		}
	}()
	return ch
}

func (v Vector) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		for el := v.l.Front(); el != nil; el = el.Next() {
			ch <- el.Value.(pair).value
		}
	}()
	return ch
}

func (a *Vector) Diff(b *Vector) []string {
	ret := make([]string, len(a.m))
	for el := range a.Iter() {
		str := el.(string)
		if !b.Has(str) {
			ret = append(ret, str)
		}
	}
	return ret
}

func (a *Vector) BiDiff(b *Vector) ([]string, []string) {
	diffOneWay := func(v1, v2 *Vector, list chan []string) {
		list <- v1.Diff(v2)
	}
	anotb, bnota := make(chan []string), make(chan []string)
	go diffOneWay(a, b, anotb)
	go diffOneWay(b, a, bnota)
	return <-anotb, <-bnota
}
