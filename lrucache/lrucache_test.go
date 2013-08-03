package lrucache

import (
	"testing"
)

const (
	get = iota
	put
)

type pair struct {
	key, value interface{}
}

var tests = []struct {
	maxSize int
	puts    []pair
	want    []pair
	onEvict func(t *testing.T) func(interface{})
}{
	{
		3,
		[]pair{{"a", 0}, {"b", 1}, {"5", 51}},
		[]pair{{"5", 51}, {"b", 1}, {"a", 0}},
		func(t *testing.T) func(interface{}) {
			return func(interface{}) {
				t.Fatal("shouldn't need to evict")
			}
		},
	},
	{
		3,
		[]pair{
			{5, 0},
			{2, 1},
			{13, 51},
			{93, 5},
			{5, 1},
			{11, 99},
			{5, 1},
			{11, 99},
			{94, 5},
		},
		[]pair{{94, 5}, {11, 99}, {5, 1}},
		func(t *testing.T) func(interface{}) {
			return func(interface{}) {}
		},
	},
}

func doPuts(c *LruCache, puts []pair) error {
	for _, kv := range puts {
		c.Put(kv.key, kv.value)
	}

	return nil
}

func TestListOrder(t *testing.T) {
	for i, test := range tests {
		cache := New(test.maxSize, test.onEvict(t))

		err := doPuts(cache, test.puts)
		if err != nil {
			t.Error(err)
			continue
		}

		j := 0
		for el := cache.lruList.Front(); el != nil; el = el.Next() {
			if j >= len(test.want) {
				t.Errorf("test %v: index out of range, test case has %v values, list has more", i, j)
				break
			}

			if el.Value == test.want[j] {
				t.Errorf("test %v: got %v, want %v", i, el.Value, test.want[j].value)
			}

			j++
		}

		if j != len(test.want) {
			t.Errorf("test %v, j should be %v, got %v", i, len(test.want), j)
		}
	}
}

func TestPutGet(t *testing.T) {
	for i, test := range tests {
		cache := New(test.maxSize, test.onEvict(t))

		err := doPuts(cache, test.puts)
		if err != nil {
			t.Error(err)
		}

		for _, kv := range test.want {
			got := cache.Get(kv.key)

			if got != kv.value {
				t.Errorf("test %v: got %v, want %v", i, got, kv.value)
			}
		}
	}
}
