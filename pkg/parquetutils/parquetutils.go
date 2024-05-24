package parquetutils

import (
	"github.com/pkg/errors"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/source"
)

// ReaderConcurrency parallel number of file readers.
var ReaderConcurrency int64 = 8

// ReadAll reads all records from the parquet file.
func ReadAll[T any](sourceFile source.ParquetFile) ([]T, error) {
	r, err := reader.NewParquetReader(sourceFile, new(T), ReaderConcurrency)
	if err != nil {
		return nil, errors.Wrap(err, "can't create parquet reader")
	}
	defer r.ReadStop()

	data := make([]T, r.GetNumRows())
	if err = r.Read(&data); err != nil {
		return nil, errors.Wrap(err, "failed to read parquet data")
	}

	return data, nil
}
