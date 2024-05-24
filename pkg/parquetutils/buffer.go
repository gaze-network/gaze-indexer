// nolint: wrapcheck
package parquetutils

import (
	"errors"
	"io"
	"sync"

	"github.com/xitongsys/parquet-go/source"
)

var (
	// Make sure Buffer implements the ParquetFile interface.
	_ source.ParquetFile = (*Buffer)(nil)

	// Make sure Buffer implements the io.WriterAt interface.
	_ io.WriterAt = (*Buffer)(nil)
)

// Buffer allows reading parquet messages from a memory buffer.
type Buffer struct {
	buf []byte
	loc int
	m   sync.Mutex
}

// NewBuffer creates a new in memory parquet buffer.
func NewBuffer() *Buffer {
	return &Buffer{buf: make([]byte, 0, 512)}
}

// NewBufferFrom creates new in memory parquet buffer from the given bytes.
// It uses the provided slice as its buffer.
func NewBufferFrom(s []byte) *Buffer {
	return &Buffer{
		buf: s,
	}
}

func (b *Buffer) Create(string) (source.ParquetFile, error) {
	return &Buffer{buf: make([]byte, 0, 512)}, nil
}

func (b *Buffer) Open(string) (source.ParquetFile, error) {
	return NewBufferFrom(b.Bytes()), nil
}

// Seek seeks in the underlying memory buffer.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	newLoc := b.loc
	switch whence {
	case io.SeekStart:
		newLoc = int(offset)
	case io.SeekCurrent:
		newLoc += int(offset)
	case io.SeekEnd:
		newLoc = len(b.buf) + int(offset)
	default:
		return int64(b.loc), errors.New("Seek: invalid whence")
	}

	if newLoc < 0 {
		return int64(b.loc), errors.New("Seek: invalid offset")
	}

	if newLoc > len(b.buf) {
		newLoc = len(b.buf)
	}

	b.loc = newLoc
	return int64(b.loc), nil
}

// Read reads data form BufferFile into p.
func (b *Buffer) Read(p []byte) (n int, err error) {
	n = copy(p, b.buf[b.loc:len(b.buf)])
	b.loc += n

	if b.loc == len(b.buf) {
		return n, io.EOF
	}

	return n, nil
}

// Write writes data from p into BufferFile.
func (b *Buffer) Write(p []byte) (n int, err error) {
	n, err = b.WriteAt(p, int64(b.loc))
	if err != nil {
		return 0, err
	}
	b.loc += n
	return
}

// WriteAt writes a slice of bytes to a buffer starting at the position provided
// The number of bytes written will be returned, or error. Can overwrite previous
// written slices if the write ats overlap.
func (b *Buffer) WriteAt(p []byte, pos int64) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	pLen := len(p)
	expLen := pos + int64(pLen)
	if int64(len(b.buf)) < expLen {
		if int64(cap(b.buf)) < expLen {
			newBuf := make([]byte, expLen)
			copy(newBuf, b.buf)
			b.buf = newBuf
		}
		b.buf = b.buf[:expLen]
	}
	copy(b.buf[pos:], p)
	return pLen, nil
}

// Close is a no-op for a memory buffer.
func (*Buffer) Close() error {
	return nil
}

// Bytes returns the underlying buffer bytes.
func (b *Buffer) Bytes() []byte {
	return b.buf
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
func (b *Buffer) Reset() {
	b.m.Lock()
	defer b.m.Unlock()
	b.buf = b.buf[:0]
	b.loc = 0
}
