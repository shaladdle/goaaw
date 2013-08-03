package local

import (
	"container/list"
)

type fileCacheElem struct {
	id   string
	size int64
}

type fileCache struct {
	lruMap  map[string]*list.Element
	lruList *list.List
	curSize int64
	maxSize int64
}

// Returns true if the cache is overfull
func (c *fileCache) Get(id string, size int64) bool {
	if c.curSize+size > c.maxSize {
		return true
	}

	if elem, exists := c.lruMap[id]; exists {
		c.lruList.MoveToFront(elem)
		return false
	}

	c.lruList.PushFront(id)

	c.curSize += size

	return c.curSize > c.maxSize
}

func (c *fileCache) Pop() []string {
	ret := []string{}

	for elem := c.lruList.Back(); c.curSize > c.maxSize; elem = elem.Next() {
		value := elem.Value.(fileCacheElem)
		c.curSize -= value.size
		ret = append(ret, value.id)

		c.lruList.Remove(elem)
		delete(c.lruMap, value.id)
	}

	return ret
}

func (c *fileCache) Contains(id string) bool {
	_, exists := c.lruMap[id]
	return exists
}

func newCache(maxSize int64) *fileCache {
	return &fileCache{
		lruMap:  make(map[string]*list.Element),
		lruList: list.New(),
		maxSize: maxSize,
		curSize: 0,
	}
}
