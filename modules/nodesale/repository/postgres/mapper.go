package postgres

import (
	"encoding/json"

	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/samber/lo"
)

func mapNodes(nodes []gen.Node) []entity.Node {
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
	})
}

func mapNodeSales(nodeSales []gen.NodeSale) []entity.NodeSale {
	return lo.Map(nodeSales, func(item gen.NodeSale, index int) entity.NodeSale {
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
	})
}

func mapNodeCountByTierIndexRows(nodeCount []gen.GetNodeCountByTierIndexRow) []datagateway.GetNodeCountByTierIndexRow {
	return lo.Map(nodeCount, func(item gen.GetNodeCountByTierIndexRow, index int) datagateway.GetNodeCountByTierIndexRow {
		return datagateway.GetNodeCountByTierIndexRow{
			TierIndex: item.TierIndex,
		}
	})
}

func mapNodeSalesEvents(events []gen.Event) []entity.NodeSaleEvent {
	return lo.Map(events, func(item gen.Event, index int) entity.NodeSaleEvent {
		var meta entity.MetadataEventPurchase
		err := json.Unmarshal(item.Metadata, &meta)
		if err != nil {
			meta = entity.MetadataEventPurchase{}
		}
		return entity.NodeSaleEvent{
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
			Metadata:       &meta,
		}
	})
}
