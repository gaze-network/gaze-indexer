package bufferpool

import (
	"bytes"
	"sync"
)

const _size = 1024 // by default, create 1 KiB buffers

// global buffer pool
var pool = &sync.Pool{
	New: func() interface{} {
		return &Buffer{
			Buffer: bytes.NewBuffer(make([]byte, 0, _size)),
		}
	},
}

type Buffer struct {
	*bytes.Buffer
	pool *sync.Pool
}

// Release returns the Buffer to its pool.
//
// Callers must not retain references to the Buffer after calling Free.
func (b *Buffer) Release() {
	b.pool.Put(b)
}

// Free returns the Buffer to its Pool (alias for Release)
//
// Callers must not retain references to the Buffer after calling Free.
func (b *Buffer) Free() {
	b.Release()
}

// TrimNewline trims any final "\n" byte from the end of the buffer.
func (b *Buffer) TrimNewline() {
	if i := b.Len() - 1; i >= 0 {
		if b.Bytes()[i] == '\n' {
			b.Truncate(i)
		}
	}
}

func Get() *Buffer {
	buf := pool.Get().(*Buffer)
	buf.Reset()
	buf.pool = pool
	return buf
}

func Put(buf *Buffer) {
	pool.Put(buf)
}
