package postgres

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
)

var _ datagateway.BRC20DataGateway = (*Repository)(nil)

// warning: GetLatestBlock currently returns a types.BlockHeader with only Height and Hash fields populated.
// This is because it is known that all usage of this function only requires these fields. In the future, we may want to populate all fields for type safety.
func (r *Repository) GetLatestBlock(ctx context.Context) (types.BlockHeader, error) {
	block, err := r.queries.GetLatestIndexedBlock(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return types.BlockHeader{}, errors.WithStack(errs.NotFound)
		}
		return types.BlockHeader{}, errors.Wrap(err, "error during query")
	}
	hash, err := chainhash.NewHashFromStr(block.Hash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to parse block hash")
	}
	return types.BlockHeader{
		Height: int64(block.Height),
		Hash:   *hash,
	}, nil
}

// GetIndexedBlockByHeight implements datagateway.BRC20DataGateway.
func (r *Repository) GetIndexedBlockByHeight(ctx context.Context, height int64) (*entity.IndexedBlock, error) {
	model, err := r.queries.GetIndexedBlockByHeight(ctx, int32(height))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.Wrap(err, "error during query")
	}

	indexedBlock, err := mapIndexedBlockModelToType(model)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse indexed block model")
	}
	return &indexedBlock, nil
}

func (r *Repository) GetProcessorStats(ctx context.Context) (*entity.ProcessorStats, error) {
	model, err := r.queries.GetLatestProcessorStats(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.WithStack(err)
	}
	stats := mapProcessorStatsModelToType(model)
	return &stats, nil
}

func (r *Repository) GetInscriptionIdsInOutPoints(ctx context.Context, outPoints []wire.OutPoint) (map[ordinals.SatPoint][]ordinals.InscriptionId, error) {
	txHashArr := lo.Map(outPoints, func(outPoint wire.OutPoint, _ int) string {
		return outPoint.Hash.String()
	})
	txOutIdxArr := lo.Map(outPoints, func(outPoint wire.OutPoint, _ int) int32 {
		return int32(outPoint.Index)
	})
	models, err := r.queries.GetInscriptionsInOutPoints(ctx, gen.GetInscriptionsInOutPointsParams{
		TxHashArr:   txHashArr,
		TxOutIdxArr: txOutIdxArr,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	inscriptionIds := make(map[ordinals.SatPoint][]ordinals.InscriptionId)
	for _, model := range models {
		// sanity check
		if !model.NewSatpointTxHash.Valid || !model.NewSatpointOutIdx.Valid || !model.NewSatpointOffset.Valid {
			return nil, errors.New("invalid satpoint: missing required satpoint fields")
		}
		txHash, err := chainhash.NewHashFromStr(model.NewSatpointTxHash.String)
		if err != nil {
			return nil, errors.Wrap(err, "invalid satpoint: cannot parse txHash")
		}
		satPoint := ordinals.SatPoint{
			OutPoint: wire.OutPoint{
				Hash:  *txHash,
				Index: uint32(model.NewSatpointOutIdx.Int32),
			},
			Offset: uint64(model.NewSatpointOffset.Int64),
		}
		inscriptionId, err := ordinals.NewInscriptionIdFromString(model.InscriptionID)
		if err != nil {
			return nil, errors.Wrap(err, "invalid inscription id")
		}
		inscriptionIds[satPoint] = append(inscriptionIds[satPoint], inscriptionId)
	}
	return inscriptionIds, nil
}

func (r *Repository) GetInscriptionEntryById(ctx context.Context, id ordinals.InscriptionId) (*ordinals.InscriptionEntry, error) {
	models, err := r.queries.GetInscriptionEntriesByIds(ctx, []string{id.String()})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(models) == 0 {
		return nil, errors.WithStack(errs.NotFound)
	}
	if len(models) > 1 {
		// sanity check
		panic("multiple inscription entries found for the same id")
	}
	inscriptionEntry, err := mapInscriptionEntryModelToType(models[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &inscriptionEntry, nil
}

func (r *Repository) CreateIndexedBlock(ctx context.Context, block *entity.IndexedBlock) error {
	params := mapIndexedBlockTypeToParams(*block)
	if err := r.queries.CreateIndexedBlock(ctx, params); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) CreateProcessorStats(ctx context.Context, stats *entity.ProcessorStats) error {
	params := mapProcessorStatsTypeToParams(*stats)
	if err := r.queries.CreateProcessorStats(ctx, params); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) CreateInscriptionEntries(ctx context.Context, blockHeight uint64, entries []*ordinals.InscriptionEntry) error {
	inscriptionEntryParams := make([]gen.CreateInscriptionEntriesParams, 0)
	for _, entry := range entries {
		params, _, err := mapInscriptionEntryTypeToParams(*entry, blockHeight)
		if err != nil {
			return errors.Wrap(err, "cannot map inscription entry to create params")
		}
		inscriptionEntryParams = append(inscriptionEntryParams, params)
	}
	results := r.queries.CreateInscriptionEntries(ctx, inscriptionEntryParams)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateInscriptionEntryStates(ctx context.Context, blockHeight uint64, entryStates []*ordinals.InscriptionEntry) error {
	inscriptionEntryStatesParams := make([]gen.CreateInscriptionEntryStatesParams, 0)
	for _, entry := range entryStates {
		_, params, err := mapInscriptionEntryTypeToParams(*entry, blockHeight)
		if err != nil {
			return errors.Wrap(err, "cannot map inscription entry to create params")
		}
		inscriptionEntryStatesParams = append(inscriptionEntryStatesParams, params)
	}
	results := r.queries.CreateInscriptionEntryStates(ctx, inscriptionEntryStatesParams)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateInscriptionTransfers(ctx context.Context, transfers []*entity.InscriptionTransfer) error {
	params := lo.Map(transfers, func(transfer *entity.InscriptionTransfer, _ int) gen.CreateInscriptionTransfersParams {
		return mapInscriptionTransferTypeToParams(*transfer)
	})
	results := r.queries.CreateInscriptionTransfers(ctx, params)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}
