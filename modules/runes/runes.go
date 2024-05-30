package runes

import (
	"context"
	"strings"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/internal/postgres"
	runesapi "github.com/gaze-network/indexer-network/modules/runes/api"
	runesdatagateway "github.com/gaze-network/indexer-network/modules/runes/datagateway"
	runespostgres "github.com/gaze-network/indexer-network/modules/runes/repository/postgres"
	runesusecase "github.com/gaze-network/indexer-network/modules/runes/usecase"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/reportingclient"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
)

func New(injector do.Injector) (indexer.IndexerWorker, error) {
	ctx := do.MustInvoke[context.Context](injector)
	conf := do.MustInvoke[config.Config](injector)
	reportingClient := do.MustInvoke[*reportingclient.ReportingClient](injector)

	var (
		runesDg       runesdatagateway.RunesDataGateway
		indexerInfoDg runesdatagateway.IndexerInfoDataGateway
	)
	var cleanupFuncs []func(context.Context) error
	switch strings.ToLower(conf.Modules.Runes.Database) {
	case "postgresql", "postgres", "pg":
		pg, err := postgres.NewPool(ctx, conf.Modules.Runes.Postgres)
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
		runesRepo := runespostgres.NewRepository(pg)
		runesDg = runesRepo
		indexerInfoDg = runesRepo
	default:
		return nil, errors.Wrapf(errs.Unsupported, "%q database for indexer is not supported", conf.Modules.Runes.Database)
	}

	var bitcoinDatasource datasources.Datasource[*types.Block]
	var bitcoinClient btcclient.Contract
	switch strings.ToLower(conf.Modules.Runes.Datasource) {
	case "bitcoin-node":
		btcClient := do.MustInvoke[*rpcclient.Client](injector)
		bitcoinNodeDatasource := datasources.NewBitcoinNode(btcClient)
		bitcoinDatasource = bitcoinNodeDatasource
		bitcoinClient = bitcoinNodeDatasource
	default:
		return nil, errors.Wrapf(errs.Unsupported, "%q datasource is not supported", conf.Modules.Runes.Datasource)
	}

	processor := NewProcessor(runesDg, indexerInfoDg, bitcoinClient, conf.Network, reportingClient, cleanupFuncs)
	if err := processor.VerifyStates(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	// Mount API
	apiHandlers := lo.Uniq(conf.Modules.Runes.APIHandlers)
	for _, handler := range apiHandlers {
		switch handler { // TODO: support more handlers (e.g. gRPC)
		case "http":
			httpServer := do.MustInvoke[*fiber.App](injector)
			runesUsecase := runesusecase.New(runesDg, bitcoinClient)
			runesHTTPHandler := runesapi.NewHTTPHandler(conf.Network, runesUsecase)
			if err := runesHTTPHandler.Mount(httpServer); err != nil {
				return nil, errors.Wrap(err, "can't mount Runes API")
			}
			logger.InfoContext(ctx, "Mounted HTTP handler")
		default:
			return nil, errors.Wrapf(errs.Unsupported, "%q API handler is not supported", handler)
		}
	}

	indexer := indexer.New(processor, bitcoinDatasource)
	return indexer, nil
}
