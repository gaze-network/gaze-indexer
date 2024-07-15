package nodesale

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/api/httphandler"
	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
)

var NODESALE_MAGIC = []byte{0x6e, 0x73, 0x6f, 0x70}

const (
	NODESALE_LASTBLOCK_DEFAULT = 846851
	Version                    = "v0.0.1-alpha"
)

func New(injector do.Injector) (indexer.IndexerWorker, error) {
	ctx := do.MustInvoke[context.Context](injector)
	conf := do.MustInvoke[config.Config](injector)

	btcClient := do.MustInvoke[*rpcclient.Client](injector)
	datasource := datasources.NewBitcoinNode(btcClient)

	pg, err := postgres.NewPool(ctx, conf.Modules.Nodesale.Postgres)
	if err != nil {
		return nil, fmt.Errorf("Can't create postgres connection : %w", err)
	}
	var cleanupFuncs []func(context.Context) error
	cleanupFuncs = append(cleanupFuncs, func(ctx context.Context) error {
		pg.Close()
		return nil
	})
	repository := repository.NewRepository(pg)

	processor := &Processor{
		datagateway:  repository,
		btcClient:    datasource,
		network:      conf.Network,
		cleanupFuncs: cleanupFuncs,
	}

	httpServer := do.MustInvoke[*fiber.App](injector)
	nodesaleHandler := httphandler.New(repository)
	if err := nodesaleHandler.Mount(httpServer); err != nil {
		return nil, fmt.Errorf("Can't mount nodesale API : %w", err)
	}
	logger.InfoContext(ctx, "Mounted nodesale HTTP handler")

	indexer := indexer.New(processor, datasource)
	logger.InfoContext(ctx, "Nodesale module started.")
	return indexer, nil
}
