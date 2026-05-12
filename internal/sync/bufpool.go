package sync

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// Get returns a reset *bytes.Buffer from the pool.
func Get() *bytes.Buffer {
	buf, ok := pool.Get().(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
	}

	buf.Reset()

	return buf
}

// Put returns a *bytes.Buffer to the pool.
// Buffers larger than 1 MiB are discarded to prevent the pool from holding
// oversized allocations.
func Put(buf *bytes.Buffer) {
	if buf.Cap() > 1<<20 {
		return
	}

	pool.Put(buf)
}
