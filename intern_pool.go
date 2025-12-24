package octypes

import (
	"container/list"
	"sync"
)

const numShards = 32

// lruEntry represents an entry in the LRU cache
type lruEntry struct {
	key   string
	value string
}

// shard represents a single shard of the intern pool
type shard struct {
	mu      sync.Mutex
	cache   map[string]*list.Element
	lru     *list.List
	maxSize int
}

// InternPool is a bounded string intern pool with LRU eviction, sharded for concurrency
type InternPool struct {
	shards [numShards]*shard
	minLen int
}

// NewInternPool creates a new bounded intern pool
func NewInternPool(maxSize, minLen int) *InternPool {
	p := &InternPool{
		minLen: minLen,
	}
	
	// Distribute max size across shards
	shardSize := maxSize / numShards
	if shardSize < 1 {
		shardSize = 1
	}
	
	for i := 0; i < numShards; i++ {
		p.shards[i] = &shard{
			cache:   make(map[string]*list.Element),
			lru:     list.New(),
			maxSize: shardSize,
		}
	}
	
	return p
}

// stringHash calculates FNV-1a hash for string to select shard
func stringHash(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// Intern returns an interned version of the string
func (p *InternPool) Intern(s string) string {
	// Short strings are not worth interning
	if len(s) < p.minLen {
		return s
	}

	// Select shard
	h := stringHash(s)
	sIdx := h % numShards
	shard := p.shards[sIdx]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Check if already in cache
	if elem, ok := shard.cache[s]; ok {
		// Move to front (most recently used)
		shard.lru.MoveToFront(elem)
		return elem.Value.(*lruEntry).value
	}

	// Add to cache
	entry := &lruEntry{key: s, value: s}
	elem := shard.lru.PushFront(entry)
	shard.cache[s] = elem

	// Evict oldest if over capacity
	if shard.lru.Len() > shard.maxSize {
		oldest := shard.lru.Back()
		if oldest != nil {
			shard.lru.Remove(oldest)
			delete(shard.cache, oldest.Value.(*lruEntry).key)
		}
	}

	return s
}

// Size returns the current size of the intern pool
func (p *InternPool) Size() int {
	total := 0
	for i := 0; i < numShards; i++ {
		shard := p.shards[i]
		shard.mu.Lock()
		total += len(shard.cache)
		shard.mu.Unlock()
	}
	return total
}

// Clear removes all entries from the intern pool
func (p *InternPool) Clear() {
	for i := 0; i < numShards; i++ {
		shard := p.shards[i]
		shard.mu.Lock()
		shard.cache = make(map[string]*list.Element)
		shard.lru = list.New()
		shard.mu.Unlock()
	}
}
