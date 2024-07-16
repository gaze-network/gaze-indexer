package postgres

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
)

type Repository struct {
	db      postgres.DB
	queries *gen.Queries
	tx      pgx.Tx
}

func NewRepository(db postgres.DB) *Repository {
	return &Repository{
		db:      db,
		queries: gen.New(db),
	}
}

func (repo *Repository) AddBlock(ctx context.Context, arg entity.Block) error {
	err := repo.queries.AddBlock(ctx, gen.AddBlockParams{
		BlockHeight: arg.BlockHeight,
		BlockHash:   arg.BlockHash,
		Module:      arg.Module,
	})
	if err != nil {
		return errors.Wrap(err, "Cannot Add block")
	}

	return nil
}

func (repo *Repository) GetBlock(ctx context.Context, blockHeight int64) (*entity.Block, error) {
	block, err := repo.queries.GetBlock(ctx, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get block")
	}
	return &entity.Block{
		BlockHeight: block.BlockHeight,
		BlockHash:   block.BlockHash,
		Module:      block.Module,
	}, nil
}

func (repo *Repository) GetLastProcessedBlock(ctx context.Context) (*entity.Block, error) {
	block, err := repo.queries.GetLastProcessedBlock(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get last processed block")
	}
	return &entity.Block{
		BlockHeight: block.BlockHeight,
		BlockHash:   block.BlockHash,
		Module:      block.Module,
	}, nil
}

func (repo *Repository) RemoveBlockFrom(ctx context.Context, fromBlock int64) (int64, error) {
	affected, err := repo.queries.RemoveBlockFrom(ctx, fromBlock)
	if err != nil {
		return 0, errors.Wrap(err, "Cannot remove blocks")
	}
	return affected, nil
}

func (repo *Repository) RemoveEventsFromBlock(ctx context.Context, fromBlock int64) (int64, error) {
	affected, err := repo.queries.RemoveEventsFromBlock(ctx, fromBlock)
	if err != nil {
		return 0, errors.Wrap(err, "Cannot remove events")
	}
	return affected, nil
}

func (repo *Repository) ClearDelegate(ctx context.Context) (int64, error) {
	affected, err := repo.queries.ClearDelegate(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "Cannot clear delegate")
	}
	return affected, nil
}

func (repo *Repository) GetNodes(ctx context.Context, arg datagateway.GetNodesParams) ([]entity.Node, error) {
	nodes, err := repo.queries.GetNodes(ctx, gen.GetNodesParams{
		SaleBlock:   arg.SaleBlock,
		SaleTxIndex: arg.SaleTxIndex,
		NodeIds:     arg.NodeIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get nodes")
	}
	return lo.Map(nodes, func(item gen.Node, index int) entity.Node {
		return entity.Node{
			SaleBlock:      item.SaleBlock,
			SaleTxIndex:    item.SaleTxIndex,
			NodeID:         item.NodeID,
			TierIndex:      item.TierIndex,
			DelegatedTo:    item.DelegatedTo,
			OwnerPublicKey: item.OwnerPublicKey,
			PurchaseTxHash: item.PurchaseTxHash,
			DelegateTxHash: item.DelegateTxHash,
		}
	}), nil
}

func (repo *Repository) AddEvent(ctx context.Context, arg datagateway.AddEventParams) error {
	err := repo.queries.AddEvent(ctx, gen.AddEventParams{
		TxHash:         arg.TxHash,
		BlockHeight:    arg.BlockHeight,
		TxIndex:        arg.TxIndex,
		WalletAddress:  arg.WalletAddress,
		Valid:          arg.Valid,
		Action:         arg.Action,
		RawMessage:     arg.RawMessage,
		ParsedMessage:  arg.ParsedMessage,
		BlockTimestamp: pgtype.Timestamp{Time: arg.BlockTimestamp.UTC(), Valid: true},
		BlockHash:      arg.BlockHash,
		Metadata:       arg.Metadata,
	})
	if err != nil {
		return errors.Wrap(err, "Cannot add event")
	}
	return nil
}

func (repo *Repository) SetDelegates(ctx context.Context, arg datagateway.SetDelegatesParams) (int64, error) {
	affected, err := repo.queries.SetDelegates(ctx, gen.SetDelegatesParams{
		SaleBlock:   arg.SaleBlock,
		SaleTxIndex: arg.SaleTxIndex,
		Delegatee:   arg.Delegatee,
		NodeIds:     arg.NodeIds,
	})
	if err != nil {
		return 0, errors.Wrap(err, "Cannot set delegate")
	}
	return affected, nil
}

func (repo *Repository) AddNodesale(ctx context.Context, arg datagateway.AddNodesaleParams) error {
	err := repo.queries.AddNodesale(ctx, gen.AddNodesaleParams{
		BlockHeight:           arg.BlockHeight,
		TxIndex:               arg.TxIndex,
		Name:                  arg.Name,
		StartsAt:              pgtype.Timestamp{Time: arg.StartsAt.UTC(), Valid: true},
		EndsAt:                pgtype.Timestamp{Time: arg.EndsAt.UTC(), Valid: true},
		Tiers:                 arg.Tiers,
		SellerPublicKey:       arg.SellerPublicKey,
		MaxPerAddress:         arg.MaxPerAddress,
		DeployTxHash:          arg.DeployTxHash,
		MaxDiscountPercentage: arg.MaxDiscountPercentage,
		SellerWallet:          arg.SellerWallet,
	})
	if err != nil {
		return errors.Wrap(err, "Cannot add nodesale")
	}
	return nil
}

func (repo *Repository) GetNodesale(ctx context.Context, arg datagateway.GetNodesaleParams) ([]entity.NodeSale, error) {
	nodesales, err := repo.queries.GetNodesale(ctx, gen.GetNodesaleParams{
		BlockHeight: arg.BlockHeight,
		TxIndex:     arg.TxIndex,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get nodesale")
	}

	return lo.Map(nodesales, func(item gen.NodeSale, index int) entity.NodeSale {
		return entity.NodeSale{
			BlockHeight:           item.BlockHeight,
			TxIndex:               item.TxIndex,
			Name:                  item.Name,
			StartsAt:              item.StartsAt.Time,
			EndsAt:                item.EndsAt.Time,
			Tiers:                 item.Tiers,
			SellerPublicKey:       item.SellerPublicKey,
			MaxPerAddress:         item.MaxPerAddress,
			DeployTxHash:          item.DeployTxHash,
			MaxDiscountPercentage: item.MaxDiscountPercentage,
			SellerWallet:          item.SellerWallet,
		}
	}), nil
}

func (repo *Repository) GetNodesByOwner(ctx context.Context, arg datagateway.GetNodesByOwnerParams) ([]entity.Node, error) {
	nodes, err := repo.queries.GetNodesByOwner(ctx, gen.GetNodesByOwnerParams{
		SaleBlock:      arg.SaleBlock,
		SaleTxIndex:    arg.SaleTxIndex,
		OwnerPublicKey: arg.OwnerPublicKey,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get nodes by owner")
	}
	return lo.Map(nodes, func(item gen.Node, index int) entity.Node {
		return entity.Node{
			SaleBlock:      item.SaleBlock,
			SaleTxIndex:    item.SaleTxIndex,
			NodeID:         item.NodeID,
			TierIndex:      item.TierIndex,
			DelegatedTo:    item.DelegatedTo,
			OwnerPublicKey: item.OwnerPublicKey,
			PurchaseTxHash: item.PurchaseTxHash,
			DelegateTxHash: item.DelegateTxHash,
		}
	}), nil
}

func (repo *Repository) AddNode(ctx context.Context, arg datagateway.AddNodeParams) error {
	err := repo.queries.AddNode(ctx, gen.AddNodeParams{
		SaleBlock:      arg.SaleBlock,
		SaleTxIndex:    arg.SaleTxIndex,
		NodeID:         arg.NodeID,
		TierIndex:      arg.TierIndex,
		DelegatedTo:    arg.DelegatedTo,
		OwnerPublicKey: arg.OwnerPublicKey,
		PurchaseTxHash: arg.PurchaseTxHash,
		DelegateTxHash: arg.DelegateTxHash,
	})
	if err != nil {
		return errors.Wrap(err, "Cannot add node")
	}
	return nil
}

func (repo *Repository) GetNodeCountByTierIndex(ctx context.Context, arg datagateway.GetNodeCountByTierIndexParams) ([]datagateway.GetNodeCountByTierIndexRow, error) {
	nodecount, err := repo.queries.GetNodeCountByTierIndex(ctx, gen.GetNodeCountByTierIndexParams{
		SaleBlock:   arg.SaleBlock,
		SaleTxIndex: arg.SaleTxIndex,
		FromTier:    arg.FromTier,
		ToTier:      arg.ToTier,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get node count by tier index")
	}

	return lo.Map(nodecount, func(item gen.GetNodeCountByTierIndexRow, index int) datagateway.GetNodeCountByTierIndexRow {
		return datagateway.GetNodeCountByTierIndexRow{
			TierIndex: item.TierIndex,
		}
	}), nil
}

func (repo *Repository) GetNodesByPubkey(ctx context.Context, arg datagateway.GetNodesByPubkeyParams) ([]entity.Node, error) {
	nodes, err := repo.queries.GetNodesByPubkey(ctx, gen.GetNodesByPubkeyParams{
		SaleBlock:      arg.SaleBlock,
		SaleTxIndex:    arg.SaleTxIndex,
		OwnerPublicKey: arg.OwnerPublicKey,
		DelegatedTo:    arg.DelegatedTo,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get nodes by public key")
	}
	return lo.Map(nodes, func(item gen.Node, index int) entity.Node {
		return entity.Node{
			SaleBlock:      item.SaleBlock,
			SaleTxIndex:    item.SaleTxIndex,
			NodeID:         item.NodeID,
			TierIndex:      item.TierIndex,
			DelegatedTo:    item.DelegatedTo,
			OwnerPublicKey: item.OwnerPublicKey,
			PurchaseTxHash: item.PurchaseTxHash,
			DelegateTxHash: item.DelegateTxHash,
		}
	}), nil
}

func (repo *Repository) GetEventsByWallet(ctx context.Context, walletAddress string) ([]entity.Event, error) {
	events, err := repo.queries.GetEventsByWallet(ctx, walletAddress)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get events by wallet")
	}
	return lo.Map(events, func(item gen.Event, index int) entity.Event {
		return entity.Event{
			TxHash:         item.TxHash,
			BlockHeight:    item.BlockHeight,
			TxIndex:        item.TxIndex,
			WalletAddress:  item.WalletAddress,
			Valid:          item.Valid,
			Action:         item.Action,
			RawMessage:     item.RawMessage,
			ParsedMessage:  item.ParsedMessage,
			BlockTimestamp: item.BlockTimestamp.Time.UTC(),
			BlockHash:      item.BlockHash,
			Metadata:       item.Metadata,
		}
	}), nil
}
