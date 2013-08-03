// Package lrucache is a map + linked list based implementation of an in
// memory LRU cache.
package lrucache

import (
	"container/list"
	"fmt"
)

type LruCache struct {
	hashMap map[interface{}]*list.Element
	lruList *list.List
	curSize int
	maxSize int
	onEvict func(value interface{})
}

// REQUIRES: c.Contains(key)
func (c *LruCache) lookup(key interface{}) *list.Element {
	el, ok := c.hashMap[key]
	if !ok {
		panic(fmt.Sprintf("tried to access element not in the cache: %v", key))
	}

	return el
}

func (c *LruCache) evictIf() {
	if c.curSize > c.maxSize {
		el := c.lruList.Back()
		c.onEvict(el.Value)
		c.lruList.Remove(el)
		c.curSize--
	}
}

func New(maxSize int, onEvict func(value interface{})) *LruCache {
	return &LruCache{
		hashMap: make(map[interface{}]*list.Element),
		lruList: list.New(),
		maxSize: maxSize,
		onEvict: onEvict,
	}
}

func (c *LruCache) Put(key, value interface{}) {
	if !c.Contains(key) {
		el := c.lruList.PushFront(value)
		c.hashMap[key] = el
		c.curSize++
	} else {
		el := c.lookup(key)
		el.Value = value
		c.lruList.MoveToFront(el)
	}

	c.evictIf()
}

func (c *LruCache) Get(key interface{}) interface{} {
	el := c.lookup(key)

	c.lruList.MoveToFront(el)

	return el.Value
}

func (c *LruCache) Contains(key interface{}) bool {
	_, ok := c.hashMap[key]
	return ok
}
