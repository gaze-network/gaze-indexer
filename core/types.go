package core

import "context"

type IndexerWorker interface {
	Run(ctx context.Context) error
}
