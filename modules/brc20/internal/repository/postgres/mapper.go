package postgres

import (
	"encoding/hex"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/repository/postgres/gen"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func decimalFromNumeric(src pgtype.Numeric) decimal.NullDecimal {
	if !src.Valid || src.NaN || src.InfinityModifier != pgtype.Finite {
		return decimal.NullDecimal{}
	}
	result := decimal.NewFromBigInt(src.Int, src.Exp)
	return decimal.NewNullDecimal(result)
}

func numericFromDecimal(src decimal.Decimal) pgtype.Numeric {
	result := pgtype.Numeric{
		Int:              src.Coefficient(),
		Exp:              src.Exponent(),
		NaN:              false,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
	return result
}

func numericFromNullDecimal(src decimal.NullDecimal) pgtype.Numeric {
	if !src.Valid {
		return pgtype.Numeric{}
	}
	return numericFromDecimal(src.Decimal)
}

func mapIndexerStatesModelToType(src gen.Brc20IndexerState) entity.IndexerState {
	var createdAt time.Time
	if src.CreatedAt.Valid {
		createdAt = src.CreatedAt.Time
	}
	return entity.IndexerState{
		ClientVersion:    src.ClientVersion,
		Network:          common.Network(src.Network),
		DBVersion:        int32(src.DbVersion),
		EventHashVersion: int32(src.EventHashVersion),
		CreatedAt:        createdAt,
	}
}

func mapIndexerStatesTypeToParams(src entity.IndexerState) gen.CreateIndexerStateParams {
	return gen.CreateIndexerStateParams{
		ClientVersion:    src.ClientVersion,
		Network:          string(src.Network),
		DbVersion:        int32(src.DBVersion),
		EventHashVersion: int32(src.EventHashVersion),
	}
}

func mapIndexedBlockModelToType(src gen.Brc20IndexedBlock) (entity.IndexedBlock, error) {
	hash, err := chainhash.NewHashFromStr(src.Hash)
	if err != nil {
		return entity.IndexedBlock{}, errors.Wrap(err, "invalid block hash")
	}
	eventHash, err := chainhash.NewHashFromStr(src.EventHash)
	if err != nil {
		return entity.IndexedBlock{}, errors.Wrap(err, "invalid event hash")
	}
	cumulativeEventHash, err := chainhash.NewHashFromStr(src.CumulativeEventHash)
	if err != nil {
		return entity.IndexedBlock{}, errors.Wrap(err, "invalid cumulative event hash")
	}
	return entity.IndexedBlock{
		Height:              uint64(src.Height),
		Hash:                *hash,
		EventHash:           *eventHash,
		CumulativeEventHash: *cumulativeEventHash,
	}, nil
}

func mapIndexedBlockTypeToParams(src entity.IndexedBlock) gen.CreateIndexedBlockParams {
	return gen.CreateIndexedBlockParams{
		Height:              int32(src.Height),
		Hash:                src.Hash.String(),
		EventHash:           src.EventHash.String(),
		CumulativeEventHash: src.CumulativeEventHash.String(),
	}
}

func mapProcessorStatsModelToType(src gen.Brc20ProcessorStat) entity.ProcessorStats {
	return entity.ProcessorStats{
		BlockHeight:             uint64(src.BlockHeight),
		CursedInscriptionCount:  uint64(src.CursedInscriptionCount),
		BlessedInscriptionCount: uint64(src.BlessedInscriptionCount),
		LostSats:                uint64(src.LostSats),
	}
}

func mapProcessorStatsTypeToParams(src entity.ProcessorStats) gen.CreateProcessorStatsParams {
	return gen.CreateProcessorStatsParams{
		BlockHeight:             int32(src.BlockHeight),
		CursedInscriptionCount:  int32(src.CursedInscriptionCount),
		BlessedInscriptionCount: int32(src.BlessedInscriptionCount),
		LostSats:                int64(src.LostSats),
	}
}

func mapTickEntryModelToType(src gen.GetTickEntriesByTicksRow) (entity.TickEntry, error) {
	deployInscriptionId, err := ordinals.NewInscriptionIdFromString(src.DeployInscriptionID)
	if err != nil {
		return entity.TickEntry{}, errors.Wrap(err, "invalid deployInscriptionId")
	}
	var completedAt time.Time
	if src.CompletedAt.Valid {
		completedAt = src.CompletedAt.Time
	}
	return entity.TickEntry{
		Tick:                src.Tick,
		OriginalTick:        src.OriginalTick,
		TotalSupply:         decimalFromNumeric(src.TotalSupply).Decimal,
		Decimals:            uint16(src.Decimals),
		LimitPerMint:        decimalFromNumeric(src.LimitPerMint).Decimal,
		IsSelfMint:          src.IsSelfMint,
		DeployInscriptionId: deployInscriptionId,
		DeployedAt:          src.DeployedAt.Time,
		DeployedAtHeight:    uint64(src.DeployedAtHeight),
		MintedAmount:        decimalFromNumeric(src.MintedAmount).Decimal,
		BurnedAmount:        decimalFromNumeric(src.BurnedAmount).Decimal,
		CompletedAt:         completedAt,
		CompletedAtHeight:   lo.Ternary(src.CompletedAtHeight.Valid, uint64(src.CompletedAtHeight.Int32), 0),
	}, nil
}

func mapTickEntryTypeToParams(src entity.TickEntry, blockHeight uint64) (gen.CreateTickEntriesParams, gen.CreateTickEntryStatesParams, error) {
	return gen.CreateTickEntriesParams{
			Tick:                src.Tick,
			OriginalTick:        src.OriginalTick,
			TotalSupply:         numericFromDecimal(src.TotalSupply),
			Decimals:            int16(src.Decimals),
			LimitPerMint:        numericFromDecimal(src.LimitPerMint),
			IsSelfMint:          src.IsSelfMint,
			DeployInscriptionID: src.DeployInscriptionId.String(),
			DeployedAt:          pgtype.Timestamp{Time: src.DeployedAt, Valid: true},
			DeployedAtHeight:    int32(src.DeployedAtHeight),
		}, gen.CreateTickEntryStatesParams{
			Tick:              src.Tick,
			BlockHeight:       int32(blockHeight),
			CompletedAt:       pgtype.Timestamp{Time: src.CompletedAt, Valid: !src.CompletedAt.IsZero()},
			CompletedAtHeight: pgtype.Int4{Int32: int32(src.CompletedAtHeight), Valid: src.CompletedAtHeight != 0},
			MintedAmount:      numericFromDecimal(src.MintedAmount),
			BurnedAmount:      numericFromDecimal(src.BurnedAmount),
		}, nil
}

func mapInscriptionEntryModelToType(src gen.GetInscriptionEntriesByIdsRow) (ordinals.InscriptionEntry, error) {
	inscriptionId, err := ordinals.NewInscriptionIdFromString(src.Id)
	if err != nil {
		return ordinals.InscriptionEntry{}, errors.Wrap(err, "invalid inscription id")
	}

	var delegate, parent *ordinals.InscriptionId
	if src.Delegate.Valid {
		delegateValue, err := ordinals.NewInscriptionIdFromString(src.Delegate.String)
		if err != nil {
			return ordinals.InscriptionEntry{}, errors.Wrap(err, "invalid delegate id")
		}
		delegate = &delegateValue
	}
	// ord 0.14.0 supports only one parent
	if len(src.Parents) > 0 {
		parentValue, err := ordinals.NewInscriptionIdFromString(src.Parents[0])
		if err != nil {
			return ordinals.InscriptionEntry{}, errors.Wrap(err, "invalid parent id")
		}
		parent = &parentValue
	}

	inscription := ordinals.Inscription{
		Content:         src.Content,
		ContentEncoding: lo.Ternary(src.ContentEncoding.Valid, src.ContentEncoding.String, ""),
		ContentType:     lo.Ternary(src.ContentType.Valid, src.ContentType.String, ""),
		Delegate:        delegate,
		Metadata:        src.Metadata,
		Metaprotocol:    lo.Ternary(src.Metaprotocol.Valid, src.Metaprotocol.String, ""),
		Parent:          parent,
		Pointer:         lo.Ternary(src.Pointer.Valid, lo.ToPtr(uint64(src.Pointer.Int64)), nil),
	}

	return ordinals.InscriptionEntry{
		Id:              inscriptionId,
		Number:          src.Number,
		SequenceNumber:  uint64(src.SequenceNumber),
		Cursed:          src.Cursed,
		CursedForBRC20:  src.CursedForBrc20,
		CreatedAt:       lo.Ternary(src.CreatedAt.Valid, src.CreatedAt.Time, time.Time{}),
		CreatedAtHeight: uint64(src.CreatedAtHeight),
		Inscription:     inscription,
		TransferCount:   lo.Ternary(src.TransferCount.Valid, uint32(src.TransferCount.Int32), 0),
	}, nil
}

func mapInscriptionEntryTypeToParams(src ordinals.InscriptionEntry, blockHeight uint64) (gen.CreateInscriptionEntriesParams, gen.CreateInscriptionEntryStatesParams, error) {
	var delegate, metaprotocol, contentEncoding, contentType pgtype.Text
	if src.Inscription.Delegate != nil {
		delegate = pgtype.Text{String: src.Inscription.Delegate.String(), Valid: true}
	}
	if src.Inscription.Metaprotocol != "" {
		metaprotocol = pgtype.Text{String: src.Inscription.Metaprotocol, Valid: true}
	}
	if src.Inscription.ContentEncoding != "" {
		contentEncoding = pgtype.Text{String: src.Inscription.ContentEncoding, Valid: true}
	}
	if src.Inscription.ContentType != "" {
		contentType = pgtype.Text{String: src.Inscription.ContentType, Valid: true}
	}
	var parents []string
	if src.Inscription.Parent != nil {
		parents = append(parents, src.Inscription.Parent.String())
	}
	var pointer pgtype.Int8
	if src.Inscription.Pointer != nil {
		pointer = pgtype.Int8{Int64: int64(*src.Inscription.Pointer), Valid: true}
	}
	return gen.CreateInscriptionEntriesParams{
			Id:              src.Id.String(),
			Number:          src.Number,
			SequenceNumber:  int64(src.SequenceNumber),
			Delegate:        delegate,
			Metadata:        src.Inscription.Metadata,
			Metaprotocol:    metaprotocol,
			Parents:         parents,
			Pointer:         pointer,
			Content:         src.Inscription.Content,
			ContentEncoding: contentEncoding,
			ContentType:     contentType,
			Cursed:          src.Cursed,
			CursedForBrc20:  src.CursedForBRC20,
			CreatedAt:       lo.Ternary(!src.CreatedAt.IsZero(), pgtype.Timestamp{Time: src.CreatedAt, Valid: true}, pgtype.Timestamp{}),
			CreatedAtHeight: int32(src.CreatedAtHeight),
		}, gen.CreateInscriptionEntryStatesParams{
			Id:            src.Id.String(),
			BlockHeight:   int32(blockHeight),
			TransferCount: int32(src.TransferCount),
		}, nil
}

func mapInscriptionTransferModelToType(src gen.GetInscriptionTransfersInOutPointsRow) (entity.InscriptionTransfer, error) {
	inscriptionId, err := ordinals.NewInscriptionIdFromString(src.InscriptionID)
	if err != nil {
		return entity.InscriptionTransfer{}, errors.Wrap(err, "invalid inscription id")
	}
	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return entity.InscriptionTransfer{}, errors.Wrap(err, "invalid tx hash")
	}
	var oldSatPoint, newSatPoint ordinals.SatPoint
	if src.OldSatpointTxHash.Valid {
		if !src.OldSatpointOutIdx.Valid || !src.OldSatpointOffset.Valid {
			return entity.InscriptionTransfer{}, errors.New("old satpoint out idx and offset must exist if hash exists")
		}
		txHash, err := chainhash.NewHashFromStr(src.OldSatpointTxHash.String)
		if err != nil {
			return entity.InscriptionTransfer{}, errors.Wrap(err, "invalid old satpoint tx hash")
		}
		oldSatPoint = ordinals.SatPoint{
			OutPoint: wire.OutPoint{
				Hash:  *txHash,
				Index: uint32(src.OldSatpointOutIdx.Int32),
			},
			Offset: uint64(src.OldSatpointOffset.Int64),
		}
	}
	if src.NewSatpointTxHash.Valid {
		if !src.NewSatpointOutIdx.Valid || !src.NewSatpointOffset.Valid {
			return entity.InscriptionTransfer{}, errors.New("new satpoint out idx and offset must exist if hash exists")
		}
		txHash, err := chainhash.NewHashFromStr(src.NewSatpointTxHash.String)
		if err != nil {
			return entity.InscriptionTransfer{}, errors.Wrap(err, "invalid new satpoint tx hash")
		}
		newSatPoint = ordinals.SatPoint{
			OutPoint: wire.OutPoint{
				Hash:  *txHash,
				Index: uint32(src.NewSatpointOutIdx.Int32),
			},
			Offset: uint64(src.NewSatpointOffset.Int64),
		}
	}
	newPkScript, err := hex.DecodeString(src.NewPkscript)
	if err != nil {
		return entity.InscriptionTransfer{}, errors.Wrap(err, "failed to parse pkscript")
	}

	return entity.InscriptionTransfer{
		InscriptionId:  inscriptionId,
		BlockHeight:    uint64(src.BlockHeight),
		TxIndex:        uint32(src.TxIndex),
		TxHash:         *txHash,
		FromInputIndex: uint32(src.FromInputIndex),
		Content:        src.Content,
		OldSatPoint:    oldSatPoint,
		NewSatPoint:    newSatPoint,
		NewPkScript:    newPkScript,
		NewOutputValue: uint64(src.NewOutputValue),
		SentAsFee:      src.SentAsFee,
		TransferCount:  uint32(src.TransferCount),
	}, nil
}

func mapInscriptionTransferTypeToParams(src entity.InscriptionTransfer) gen.CreateInscriptionTransfersParams {
	return gen.CreateInscriptionTransfersParams{
		InscriptionID:     src.InscriptionId.String(),
		BlockHeight:       int32(src.BlockHeight),
		TxIndex:           int32(src.TxIndex),
		TxHash:            src.TxHash.String(),
		FromInputIndex:    int32(src.FromInputIndex),
		OldSatpointTxHash: lo.Ternary(src.OldSatPoint != ordinals.SatPoint{}, pgtype.Text{String: src.OldSatPoint.OutPoint.Hash.String(), Valid: true}, pgtype.Text{}),
		OldSatpointOutIdx: lo.Ternary(src.OldSatPoint != ordinals.SatPoint{}, pgtype.Int4{Int32: int32(src.OldSatPoint.OutPoint.Index), Valid: true}, pgtype.Int4{}),
		OldSatpointOffset: lo.Ternary(src.OldSatPoint != ordinals.SatPoint{}, pgtype.Int8{Int64: int64(src.OldSatPoint.Offset), Valid: true}, pgtype.Int8{}),
		NewSatpointTxHash: lo.Ternary(src.NewSatPoint != ordinals.SatPoint{}, pgtype.Text{String: src.NewSatPoint.OutPoint.Hash.String(), Valid: true}, pgtype.Text{}),
		NewSatpointOutIdx: lo.Ternary(src.NewSatPoint != ordinals.SatPoint{}, pgtype.Int4{Int32: int32(src.NewSatPoint.OutPoint.Index), Valid: true}, pgtype.Int4{}),
		NewSatpointOffset: lo.Ternary(src.NewSatPoint != ordinals.SatPoint{}, pgtype.Int8{Int64: int64(src.NewSatPoint.Offset), Valid: true}, pgtype.Int8{}),
		NewPkscript:       hex.EncodeToString(src.NewPkScript),
		NewOutputValue:    int64(src.NewOutputValue),
		SentAsFee:         src.SentAsFee,
		TransferCount:     int32(src.TransferCount),
	}
}

func mapEventDeployModelToType(src gen.Brc20EventDeploy) (entity.EventDeploy, error) {
	inscriptionId, err := ordinals.NewInscriptionIdFromString(src.InscriptionID)
	if err != nil {
		return entity.EventDeploy{}, errors.Wrap(err, "invalid inscription id")
	}
	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return entity.EventDeploy{}, errors.Wrap(err, "invalid tx hash")
	}
	pkScript, err := hex.DecodeString(src.Pkscript)
	if err != nil {
		return entity.EventDeploy{}, errors.Wrap(err, "failed to parse pkscript")
	}
	satPoint, err := ordinals.NewSatPointFromString(src.Satpoint)
	if err != nil {
		return entity.EventDeploy{}, errors.Wrap(err, "cannot parse satpoint")
	}
	return entity.EventDeploy{
		Id:                uint64(src.Id),
		InscriptionId:     inscriptionId,
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            *txHash,
		BlockHeight:       uint64(src.BlockHeight),
		TxIndex:           uint32(src.TxIndex),
		Timestamp:         src.Timestamp.Time,
		PkScript:          pkScript,
		SatPoint:          satPoint,
		TotalSupply:       decimalFromNumeric(src.TotalSupply).Decimal,
		Decimals:          uint16(src.Decimals),
		LimitPerMint:      decimalFromNumeric(src.LimitPerMint).Decimal,
		IsSelfMint:        src.IsSelfMint,
	}, nil
}

func mapEventDeployTypeToParams(src entity.EventDeploy) (gen.CreateEventDeploysParams, error) {
	var timestamp pgtype.Timestamp
	if !src.Timestamp.IsZero() {
		timestamp = pgtype.Timestamp{Time: src.Timestamp, Valid: true}
	}
	return gen.CreateEventDeploysParams{
		InscriptionID:     src.InscriptionId.String(),
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            src.TxHash.String(),
		BlockHeight:       int32(src.BlockHeight),
		TxIndex:           int32(src.TxIndex),
		Timestamp:         timestamp,
		Pkscript:          hex.EncodeToString(src.PkScript),
		Satpoint:          src.SatPoint.String(),
		TotalSupply:       numericFromDecimal(src.TotalSupply),
		Decimals:          int16(src.Decimals),
		LimitPerMint:      numericFromDecimal(src.LimitPerMint),
		IsSelfMint:        src.IsSelfMint,
	}, nil
}

func mapEventMintModelToType(src gen.Brc20EventMint) (entity.EventMint, error) {
	inscriptionId, err := ordinals.NewInscriptionIdFromString(src.InscriptionID)
	if err != nil {
		return entity.EventMint{}, errors.Wrap(err, "invalid inscription id")
	}
	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return entity.EventMint{}, errors.Wrap(err, "invalid tx hash")
	}
	pkScript, err := hex.DecodeString(src.Pkscript)
	if err != nil {
		return entity.EventMint{}, errors.Wrap(err, "failed to parse pkscript")
	}
	satPoint, err := ordinals.NewSatPointFromString(src.Satpoint)
	if err != nil {
		return entity.EventMint{}, errors.Wrap(err, "cannot parse satpoint")
	}
	var parentId *ordinals.InscriptionId
	if src.ParentID.Valid {
		parentIdValue, err := ordinals.NewInscriptionIdFromString(src.ParentID.String)
		if err != nil {
			return entity.EventMint{}, errors.Wrap(err, "invalid parent id")
		}
		parentId = &parentIdValue
	}
	return entity.EventMint{
		Id:                uint64(src.Id),
		InscriptionId:     inscriptionId,
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            *txHash,
		BlockHeight:       uint64(src.BlockHeight),
		TxIndex:           uint32(src.TxIndex),
		Timestamp:         src.Timestamp.Time,
		PkScript:          pkScript,
		SatPoint:          satPoint,
		Amount:            decimalFromNumeric(src.Amount).Decimal,
		ParentId:          parentId,
	}, nil
}

func mapEventMintTypeToParams(src entity.EventMint) (gen.CreateEventMintsParams, error) {
	var timestamp pgtype.Timestamp
	if !src.Timestamp.IsZero() {
		timestamp = pgtype.Timestamp{Time: src.Timestamp, Valid: true}
	}
	var parentId pgtype.Text
	if src.ParentId != nil {
		parentId = pgtype.Text{String: src.ParentId.String(), Valid: true}
	}
	return gen.CreateEventMintsParams{
		InscriptionID:     src.InscriptionId.String(),
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            src.TxHash.String(),
		BlockHeight:       int32(src.BlockHeight),
		TxIndex:           int32(src.TxIndex),
		Timestamp:         timestamp,
		Pkscript:          hex.EncodeToString(src.PkScript),
		Satpoint:          src.SatPoint.String(),
		Amount:            numericFromDecimal(src.Amount),
		ParentID:          parentId,
	}, nil
}

func mapEventInscribeTransferModelToType(src gen.Brc20EventInscribeTransfer) (entity.EventInscribeTransfer, error) {
	inscriptionId, err := ordinals.NewInscriptionIdFromString(src.InscriptionID)
	if err != nil {
		return entity.EventInscribeTransfer{}, errors.Wrap(err, "cannot parse inscription id")
	}
	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return entity.EventInscribeTransfer{}, errors.Wrap(err, "cannot parse hash")
	}
	pkScript, err := hex.DecodeString(src.Pkscript)
	if err != nil {
		return entity.EventInscribeTransfer{}, errors.Wrap(err, "cannot parse pkScript")
	}
	satPoint, err := ordinals.NewSatPointFromString(src.Satpoint)
	if err != nil {
		return entity.EventInscribeTransfer{}, errors.Wrap(err, "cannot parse satPoint")
	}
	return entity.EventInscribeTransfer{
		Id:                uint64(src.Id),
		InscriptionId:     inscriptionId,
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            *txHash,
		BlockHeight:       uint64(src.BlockHeight),
		TxIndex:           uint32(src.TxIndex),
		Timestamp:         src.Timestamp.Time,
		PkScript:          pkScript,
		SatPoint:          satPoint,
		OutputIndex:       uint32(src.OutputIndex),
		SatsAmount:        uint64(src.SatsAmount),
		Amount:            decimalFromNumeric(src.Amount).Decimal,
	}, nil
}

func mapEventInscribeTransferTypeToParams(src entity.EventInscribeTransfer) (gen.CreateEventInscribeTransfersParams, error) {
	var timestamp pgtype.Timestamp
	if !src.Timestamp.IsZero() {
		timestamp = pgtype.Timestamp{Time: src.Timestamp, Valid: true}
	}
	return gen.CreateEventInscribeTransfersParams{
		InscriptionID:     src.InscriptionId.String(),
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            src.TxHash.String(),
		BlockHeight:       int32(src.BlockHeight),
		TxIndex:           int32(src.TxIndex),
		Timestamp:         timestamp,
		Pkscript:          hex.EncodeToString(src.PkScript),
		Satpoint:          src.SatPoint.String(),
		OutputIndex:       int32(src.OutputIndex),
		SatsAmount:        int64(src.SatsAmount),
		Amount:            numericFromDecimal(src.Amount),
	}, nil
}

func mapEventTransferTransferModelToType(src gen.Brc20EventTransferTransfer) (entity.EventTransferTransfer, error) {
	inscriptionId, err := ordinals.NewInscriptionIdFromString(src.InscriptionID)
	if err != nil {
		return entity.EventTransferTransfer{}, errors.Wrap(err, "cannot parse inscription id")
	}
	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return entity.EventTransferTransfer{}, errors.Wrap(err, "cannot parse hash")
	}
	fromPkScript, err := hex.DecodeString(src.FromPkscript)
	if err != nil {
		return entity.EventTransferTransfer{}, errors.Wrap(err, "cannot parse fromPkScript")
	}
	fromSatPoint, err := ordinals.NewSatPointFromString(src.FromSatpoint)
	if err != nil {
		return entity.EventTransferTransfer{}, errors.Wrap(err, "cannot parse fromSatPoint")
	}
	toPkScript, err := hex.DecodeString(src.ToPkscript)
	if err != nil {
		return entity.EventTransferTransfer{}, errors.Wrap(err, "cannot parse toPkScript")
	}
	toSatPoint, err := ordinals.NewSatPointFromString(src.ToSatpoint)
	if err != nil {
		return entity.EventTransferTransfer{}, errors.Wrap(err, "cannot parse toSatPoint")
	}
	return entity.EventTransferTransfer{
		Id:                uint64(src.Id),
		InscriptionId:     inscriptionId,
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            *txHash,
		BlockHeight:       uint64(src.BlockHeight),
		TxIndex:           uint32(src.TxIndex),
		Timestamp:         src.Timestamp.Time,
		FromPkScript:      fromPkScript,
		FromSatPoint:      fromSatPoint,
		FromInputIndex:    uint32(src.FromInputIndex),
		ToPkScript:        toPkScript,
		ToSatPoint:        toSatPoint,
		ToOutputIndex:     uint32(src.ToOutputIndex),
		SpentAsFee:        src.SpentAsFee,
		Amount:            decimalFromNumeric(src.Amount).Decimal,
	}, nil
}

func mapEventTransferTransferTypeToParams(src entity.EventTransferTransfer) (gen.CreateEventTransferTransfersParams, error) {
	var timestamp pgtype.Timestamp
	if !src.Timestamp.IsZero() {
		timestamp = pgtype.Timestamp{Time: src.Timestamp, Valid: true}
	}
	return gen.CreateEventTransferTransfersParams{
		InscriptionID:     src.InscriptionId.String(),
		InscriptionNumber: src.InscriptionNumber,
		Tick:              src.Tick,
		OriginalTick:      src.OriginalTick,
		TxHash:            src.TxHash.String(),
		BlockHeight:       int32(src.BlockHeight),
		TxIndex:           int32(src.TxIndex),
		Timestamp:         timestamp,
		FromPkscript:      hex.EncodeToString(src.FromPkScript),
		FromSatpoint:      src.FromSatPoint.String(),
		FromInputIndex:    int32(src.FromInputIndex),
		ToPkscript:        hex.EncodeToString(src.ToPkScript),
		ToSatpoint:        src.ToSatPoint.String(),
		ToOutputIndex:     int32(src.ToOutputIndex),
		SpentAsFee:        src.SpentAsFee,
		Amount:            numericFromDecimal(src.Amount),
	}, nil
}

func mapBalanceModelToType(src gen.Brc20Balance) (entity.Balance, error) {
	pkScript, err := hex.DecodeString(src.Pkscript)
	if err != nil {
		return entity.Balance{}, errors.Wrap(err, "failed to parse pkscript")
	}
	return entity.Balance{
		PkScript:         pkScript,
		Tick:             src.Tick,
		BlockHeight:      uint64(src.BlockHeight),
		OverallBalance:   decimalFromNumeric(src.OverallBalance).Decimal,
		AvailableBalance: decimalFromNumeric(src.AvailableBalance).Decimal,
	}, nil
}
