package brc20

import (
	"context"
	"strings"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	brc20postgres "github.com/gaze-network/indexer-network/modules/brc20/internal/repository/postgres"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)

func New(injector do.Injector) (indexer.IndexerWorker, error) {
	ctx := do.MustInvoke[context.Context](injector)
	conf := do.MustInvoke[config.Config](injector)
	// reportingClient := do.MustInvoke[*reportingclient.ReportingClient](injector)

	cleanupFuncs := make([]func(context.Context) error, 0)
	var brc20Dg datagateway.BRC20DataGateway
	var indexerInfoDg datagateway.IndexerInfoDataGateway
	switch strings.ToLower(conf.Modules.BRC20.Database) {
	case "postgresql", "postgres", "pg":
		pg, err := postgres.NewPool(ctx, conf.Modules.BRC20.Postgres)
		if err != nil {
			if errors.Is(err, errs.InvalidArgument) {
				return nil, errors.Wrap(err, "Invalid Postgres configuration for indexer")
			}
			return nil, errors.Wrap(err, "can't create Postgres connection pool")
		}
		cleanupFuncs = append(cleanupFuncs, func(ctx context.Context) error {
			pg.Close()
			return nil
		})
		brc20Repo := brc20postgres.NewRepository(pg)
		brc20Dg = brc20Repo
		indexerInfoDg = brc20Repo
	default:
		return nil, errors.Wrapf(errs.Unsupported, "%q database for indexer is not supported", conf.Modules.BRC20.Database)
	}

	var bitcoinDatasource datasources.Datasource[*types.Block]
	var bitcoinClient btcclient.Contract
	switch strings.ToLower(conf.Modules.BRC20.Datasource) {
	case "bitcoin-node":
		btcClient := do.MustInvoke[*rpcclient.Client](injector)
		bitcoinNodeDatasource := datasources.NewBitcoinNode(btcClient)
		bitcoinDatasource = bitcoinNodeDatasource
		bitcoinClient = bitcoinNodeDatasource
	default:
		return nil, errors.Wrapf(errs.Unsupported, "%q datasource is not supported", conf.Modules.BRC20.Datasource)
	}

	processor, err := NewProcessor(brc20Dg, indexerInfoDg, bitcoinClient, conf.Network, 2, cleanupFuncs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := processor.VerifyStates(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	indexer := indexer.New(processor, bitcoinDatasource)
	return indexer, nil
}
