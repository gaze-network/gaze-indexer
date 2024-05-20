// nolint: wrapcheck
package parquetutils

import (
	parquetbuffer "github.com/xitongsys/parquet-go-source/buffer"
	"github.com/xitongsys/parquet-go/source"
)

// BufferFile allows reading parquet messages from a memory buffer.
type BufferFile struct {
	underlying *parquetbuffer.BufferFile
}

// NewBufferFile creates new in memory parquet buffer from the given bytes.
// It uses the provided slice as its buffer.
func NewBufferFile(s []byte) *BufferFile {
	return &BufferFile{
		underlying: parquetbuffer.NewBufferFileFromBytesNoAlloc(s),
	}
}

func (bf BufferFile) Create(string) (source.ParquetFile, error) {
	return &BufferFile{
		underlying: parquetbuffer.NewBufferFile(),
	}, nil
}

func (bf BufferFile) Open(string) (source.ParquetFile, error) {
	return NewBufferFile(bf.Bytes()), nil
}

// Seek seeks in the underlying memory buffer.
func (bf *BufferFile) Seek(offset int64, whence int) (int64, error) {
	return bf.underlying.Seek(offset, whence)
}

// Read reads data form BufferFile into p.
func (bf *BufferFile) Read(p []byte) (n int, err error) {
	return bf.underlying.Read(p)
}

// Write writes data from p into BufferFile.
func (bf *BufferFile) Write(p []byte) (n int, err error) {
	return bf.underlying.Write(p)
}

// Close is a no-op for a memory buffer.
func (bf BufferFile) Close() error {
	return bf.underlying.Close()
}

func (bf BufferFile) Bytes() []byte {
	return bf.underlying.Bytes()
}
