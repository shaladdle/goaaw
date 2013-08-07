package blkstore

import (
	"bytes"
	"strconv"
	"testing"

	"aaw/testutil"
)

func newMemCache(maxSize int64) BlkCache {
	return NewCache(NewMemStore(), maxSize)
}

type kvpair struct {
	key   string
	value []byte
}

func genRandItems(t *testing.T, numItems, itemSize int64) []kvpair {
	items := make([]kvpair, numItems)
	for i := range items {
		buf := &bytes.Buffer{}
		if err := testutil.WriteRandFile(buf, itemSize); err != nil {
			t.Fatal(err)
		}

		items[i] = kvpair{
			key:   strconv.Itoa(i),
			value: buf.Bytes(),
		}
	}

	return items
}

func putItems(t *testing.T, cache BlkCache, items []kvpair) {
	for _, item := range items {
		if err := cache.Put(item.key, item.value); err != nil {
			t.Errorf("put error: %v", err)
		}
	}
}

func getAndCheckItems(t *testing.T, cache BlkCache, items []kvpair) {
	for i, item := range items {
		if b, err := cache.Get(item.key); err != nil {
			t.Errorf("get error: %v", err)
		} else if !bytes.Equal(b, item.value) {
			t.Errorf("bytes for item %v not equal", i)
		}
	}
}

func TestCacheGetPut(t *testing.T) {
	const (
		numItems = 10
		itemSize = testutil.KB
		maxSize  = numItems * itemSize
	)

	items := genRandItems(t, numItems, itemSize)
	cache := newMemCache(maxSize)

	putItems(t, cache, items)
	getAndCheckItems(t, cache, items)
}

func TestCacheSizeLimit(t *testing.T) {
	const (
		numItems = 20
		itemSize = testutil.KB
		maxSize  = numItems * itemSize
	)

	items := genRandItems(t, numItems, itemSize)
	cache := newMemCache(maxSize)

	putItems(t, cache, items)

	if curSize := cache.(*blkcache).curSize; curSize > maxSize {
		t.Fatal("cache size is greater than maxSize: %v > %v", curSize, maxSize)
	}
}

func TestCachePutTooBig(t *testing.T) {
	const maxSize = testutil.KB

	cache := newMemCache(maxSize)
	if err := cache.Put("testitem", make([]byte, maxSize+1)); err == nil {
		t.Errorf("put should error, but did not", err)
	}
}
