package blkstore

import (
	"container/list"
	"fmt"
)

type blkcache struct {
	store   BlkStore
	lruList *list.List
	lruMap  map[string]*list.Element
	maxSize int64
	curSize int64
}

type blkcacheListEl struct {
	key  string
	size int64
}

func NewCache(store BlkStore, maxSize int64) BlkCache {
	return &blkcache{
		store:   store,
		lruList: list.New(),
		lruMap:  make(map[string]*list.Element),
		maxSize: maxSize,
	}
}

func (c *blkcache) Get(key string) ([]byte, error) {
	if _, ok := c.lruMap[key]; !ok {
		return nil, fmt.Errorf("key '%v' is not in the blkcache")
	}

	b, err := c.store.Get(key)
	if err != nil {
		return nil, err
	}

	c.moveToFront(key)

	return b, nil
}

func (c *blkcache) Put(key string, value []byte) error {
	size := int64(len(value))
	if size > c.maxSize {
		return fmt.Errorf("value for key '%v' is larger than the max blkcache size, max: %v, len(value): %v", key, c.maxSize, size)
	}

	// If the item is already in the blkcache, evict it since we will replace it
	// with the new one.
	if el, ok := c.lruMap[key]; ok {
		if err := c.evictEl(el); err != nil {
			return err
		}
	}

	// Evict down to the size limit.
	for size+c.curSize > c.maxSize {
		if err := c.evictEl(c.lruList.Back()); err != nil {
			return err
		}
	}

	// Actually put the item in the blkcache.
	if err := c.store.Put(key, value); err != nil {
		return err
	}

	// Add the item to the blkcache data structures.
	c.curSize += size
	c.lruMap[key] = c.lruList.PushFront(blkcacheListEl{key, size})

	return nil
}

func (c *blkcache) Has(key string) bool {
	_, has := c.lruMap[key]
	return has
}

func (c *blkcache) moveToFront(key string) {
	el := c.lruMap[key]
	c.lruList.MoveToFront(el)
}

func (c *blkcache) evictEl(el *list.Element) error {
	info := el.Value.(blkcacheListEl)

	if err := c.store.Delete(info.key); err != nil {
		return err
	}

	c.lruList.Remove(el)
	delete(c.lruMap, info.key)
	c.curSize -= info.size

	return nil
}
