package hackernews

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachePutAndGet(t *testing.T) {
	cache := EntryCache{}
	entry1 := Entry{Title: "entry1", ID: 1}
	entry2 := Entry{Title: "entry2", ID: 2}
	cache.Put(entry1)
	cache.Put(entry2)
	assert.Equal(t, entry1, *cache.Get(1))
	assert.Equal(t, entry2, *cache.Get(2))

	assert.NotEqual(t, entry2, *cache.Get(1))
}

func TestCacheGarbageCollect(t *testing.T) {
	entry1 := Entry{Title: "entry1", ID: 1}
	entry2 := Entry{Title: "entry2", ID: 2}
	entry3 := Entry{Title: "entry3", ID: 3}
	entry4 := Entry{Title: "entry4", ID: 4}

	cache := EntryCache{}
	cache.Put(entry1)
	cache.Put(entry2)
	cache.Put(entry3)
	cache.GC()

	assert.NotNil(t, cache.Get(1))
	assert.NotNil(t, cache.Get(2))
	cache.GC()

	assert.NotNil(t, cache.Get(1))
	assert.Nil(t, cache.Get(3))
	cache.Put(entry4)
	cache.GC()

	assert.Nil(t, cache.Get(2))
	assert.NotNil(t, cache.Get(4))
	cache.GC()

	assert.Nil(t, cache.Get(1))
	assert.Nil(t, cache.Get(2))
	assert.Nil(t, cache.Get(3))
	assert.NotNil(t, cache.Get(4))
}
