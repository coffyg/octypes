package octypes

import "sync"

// MaxPoolBufferSize is the maximum size of buffers that will be returned to the pool
// Buffers larger than this will be discarded to prevent unbounded memory growth
const MaxPoolBufferSize = 1024 * 1024 // 1MB

// putBufferSafe returns a buffer to the pool only if it's not too large
// This prevents the pool from accumulating very large buffers
func putBufferSafe(pool *sync.Pool, buf []byte) {
	// Only return buffers that aren't too large
	if cap(buf) <= MaxPoolBufferSize {
		pool.Put(buf)
	}
	// Otherwise, let the buffer be garbage collected
}