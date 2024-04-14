package datasources

import (
	"context"
)

type Datasource[T any] interface {
	Name() string
	Fetch(ctx context.Context, from, to int64) (T, error)
	FetchAsync(ctx context.Context, from, to int64, ch chan<- T) (*ClientSubscription[T], error)
}
