package octypes

import (
	"container/list"
	"sync"
)

// lruEntry represents an entry in the LRU cache
type lruEntry struct {
	key   string
	value string
}

// InternPool is a bounded string intern pool with LRU eviction
type InternPool struct {
	mu       sync.Mutex
	cache    map[string]*list.Element
	lru      *list.List
	maxSize  int
	minLen   int
}

// NewInternPool creates a new bounded intern pool
func NewInternPool(maxSize, minLen int) *InternPool {
	return &InternPool{
		cache:   make(map[string]*list.Element),
		lru:     list.New(),
		maxSize: maxSize,
		minLen:  minLen,
	}
}

// Intern returns an interned version of the string
func (p *InternPool) Intern(s string) string {
	// Short strings are not worth interning
	if len(s) < p.minLen {
		return s
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already in cache
	if elem, ok := p.cache[s]; ok {
		// Move to front (most recently used)
		p.lru.MoveToFront(elem)
		return elem.Value.(*lruEntry).value
	}

	// Add to cache
	entry := &lruEntry{key: s, value: s}
	elem := p.lru.PushFront(entry)
	p.cache[s] = elem

	// Evict oldest if over capacity
	if p.lru.Len() > p.maxSize {
		oldest := p.lru.Back()
		if oldest != nil {
			p.lru.Remove(oldest)
			delete(p.cache, oldest.Value.(*lruEntry).key)
		}
	}

	return s
}

// Size returns the current size of the intern pool
func (p *InternPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.cache)
}

// Clear removes all entries from the intern pool
func (p *InternPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache = make(map[string]*list.Element)
	p.lru = list.New()
}