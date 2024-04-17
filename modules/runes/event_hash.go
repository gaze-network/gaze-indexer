package runes

import (
	"bytes"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

// TODO: implement test to ensure that the event hash is calculated the same way for same version
func (p *Processor) calculateEventHash(header types.BlockHeader) (chainhash.Hash, error) {
	payload, err := p.getHashPayload(header)
	if err != nil {
		return chainhash.Hash{}, errors.Wrap(err, "failed to get hash payload")
	}
	return chainhash.DoubleHashH(payload), nil
}

func (p *Processor) getHashPayload(header types.BlockHeader) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("payload:v" + strconv.Itoa(EventHashVersion) + ":")
	sb.WriteString("blockHash:")
	sb.Write(header.Hash[:])

	// serialize new rune entries
	{
		runeEntries := lo.Values(p.newRuneEntries)
		slices.SortFunc(runeEntries, func(t1, t2 *runes.RuneEntry) int {
			return int(t1.Number) - int(t2.Number)
		})
		for _, entry := range runeEntries {
			sb.Write(serializeNewRuneEntry(entry))
		}
	}
	// serialize new rune entry states
	{
		runeIds := lo.Keys(p.newRuneEntryStates)
		slices.SortFunc(runeIds, func(t1, t2 runes.RuneId) int {
			return t1.Cmp(t2)
		})
		for _, runeId := range runeIds {
			sb.Write(serializeNewRuneEntryState(p.newRuneEntryStates[runeId]))
		}
	}
	// serialize new out point balances
	sb.Write(serializeNewOutPointBalances(p.newOutPointBalances))

	// serialize spend out points
	sb.Write(serializeSpendOutPoints(p.newSpendOutPoints))

	// serialize new balances
	{
		bytes, err := serializeNewBalances(p.newBalances)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize new balances")
		}
		sb.Write(bytes)
	}

	// serialize new txs
	// sort txs by block height and index
	{
		bytes, err := serializeRuneTxs(p.newRuneTxs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize new rune txs")
		}
		sb.Write(bytes)
	}
	return []byte(sb.String()), nil
}

func serializeNewRuneEntry(entry *runes.RuneEntry) []byte {
	var sb strings.Builder
	sb.WriteString("newRuneEntry:")
	sb.WriteString("runeId:" + entry.RuneId.String())
	sb.WriteString("number:" + strconv.Itoa(int(entry.Number)))
	sb.WriteString("divisibility:" + strconv.Itoa(int(entry.Divisibility)))
	sb.WriteString("premine:" + entry.Premine.String())
	sb.WriteString("rune:" + entry.SpacedRune.Rune.String())
	sb.WriteString("spacers:" + strconv.Itoa(int(entry.SpacedRune.Spacers)))
	sb.WriteString("symbol:" + string(entry.Symbol))
	if entry.Terms != nil {
		sb.WriteString("terms:")
		terms := entry.Terms
		if terms.Amount != nil {
			sb.WriteString("amount:" + terms.Amount.String())
		}
		if terms.Cap != nil {
			sb.WriteString("cap:" + terms.Cap.String())
		}
		if terms.HeightStart != nil {
			sb.WriteString("heightStart:" + strconv.Itoa(int(*terms.HeightStart)))
		}
		if terms.HeightEnd != nil {
			sb.WriteString("heightEnd:" + strconv.Itoa(int(*terms.HeightEnd)))
		}
		if terms.OffsetStart != nil {
			sb.WriteString("offsetStart:" + strconv.Itoa(int(*terms.OffsetStart)))
		}
		if terms.OffsetEnd != nil {
			sb.WriteString("offsetEnd:" + strconv.Itoa(int(*terms.OffsetEnd)))
		}
	}
	sb.WriteString("turbo:" + strconv.FormatBool(entry.Turbo))
	sb.WriteString("etchingBlock:" + strconv.Itoa(int(entry.EtchingBlock)))
	sb.WriteString("etchingTxHash:" + entry.EtchingTxHash.String())
	sb.WriteString("etchedAt:" + strconv.Itoa(int(entry.EtchedAt.Unix())))
	sb.WriteString(";")
	return []byte(sb.String())
}

func serializeNewRuneEntryState(entry *runes.RuneEntry) []byte {
	var sb strings.Builder
	sb.WriteString("newRuneEntryState:")
	// write only mutable states
	sb.WriteString("runeId:" + entry.RuneId.String())
	sb.WriteString("mints:" + entry.Mints.String())
	sb.WriteString("burnedAmount:" + entry.BurnedAmount.String())
	if entry.CompletedAtHeight != nil {
		sb.WriteString("completedAtHeight:" + strconv.Itoa(int(*entry.CompletedAtHeight)))
		sb.WriteString("completedAt:" + strconv.Itoa(int(entry.CompletedAt.Unix())))
	}
	sb.WriteString(";")
	return []byte(sb.String())
}

func serializeNewOutPointBalances(outPointBalances map[wire.OutPoint]map[runes.RuneId]uint128.Uint128) []byte {
	var sb strings.Builder
	sb.WriteString("newOutPointBalances:")

	outPoints := lo.Keys(outPointBalances)
	// sort outpoints to ensure order
	slices.SortFunc(outPoints, func(t1, t2 wire.OutPoint) int {
		if t1.Hash != t2.Hash {
			return bytes.Compare(t1.Hash[:], t2.Hash[:])
		}
		return int(t1.Index) - int(t2.Index)
	})
	for _, outPoint := range outPoints {
		runeIds := lo.Keys(outPointBalances[outPoint])
		// sort runeIds to ensure order
		slices.SortFunc(runeIds, func(t1, t2 runes.RuneId) int {
			return t1.Cmp(t2)
		})
		for _, runeId := range runeIds {
			sb.WriteString("outPoint:")
			sb.WriteString("hash:")
			sb.Write(outPoint.Hash[:])
			sb.WriteString("index:" + strconv.Itoa(int(outPoint.Index)))
			sb.WriteString("runeId:" + runeId.String())
			sb.WriteString("amount:" + outPointBalances[outPoint][runeId].String())
			sb.WriteString(";")
		}
	}
	return []byte(sb.String())
}

func serializeSpendOutPoints(spendOutPoints []wire.OutPoint) []byte {
	var sb strings.Builder
	sb.WriteString("spendOutPoints:")
	// sort outpoints to ensure order
	slices.SortFunc(spendOutPoints, func(t1, t2 wire.OutPoint) int {
		if t1.Hash != t2.Hash {
			return bytes.Compare(t1.Hash[:], t2.Hash[:])
		}
		return int(t1.Index) - int(t2.Index)
	})
	for _, outPoint := range spendOutPoints {
		sb.WriteString("hash:")
		sb.Write(outPoint.Hash[:])
		sb.WriteString("index:" + strconv.Itoa(int(outPoint.Index)))
		sb.WriteString(";")
	}
	return []byte(sb.String())
}

func serializeNewBalances(balances map[string]map[runes.RuneId]uint128.Uint128) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("newBalances:")

	pkScriptStrs := lo.Keys(balances)
	// sort pkScripts to ensure order
	slices.SortFunc(pkScriptStrs, func(t1, t2 string) int {
		return strings.Compare(t1, t2)
	})
	for _, pkScriptStr := range pkScriptStrs {
		runeIds := lo.Keys(balances[pkScriptStr])
		// sort runeIds to ensure order
		slices.SortFunc(runeIds, func(t1, t2 runes.RuneId) int {
			return t1.Cmp(t2)
		})
		pkScript, err := hex.DecodeString(pkScriptStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode pkScript")
		}
		for _, runeId := range runeIds {
			sb.WriteString("pkScript:")
			sb.Write(pkScript)
			sb.WriteString("runeId:" + runeId.String())
			sb.WriteString("amount:" + balances[pkScriptStr][runeId].String())
			sb.WriteString(";")
		}
	}
	return []byte(sb.String()), nil
}

func serializeRuneTxs(txs []*entity.RuneTransaction) ([]byte, error) {
	var sb strings.Builder

	slices.SortFunc(txs, func(t1, t2 *entity.RuneTransaction) int {
		if t1.BlockHeight != t2.BlockHeight {
			return int(t1.BlockHeight) - int(t2.BlockHeight)
		}
		return int(t1.Index) - int(t2.Index)
	})

	sb.WriteString("txs:")
	for _, tx := range txs {
		sb.WriteString("hash:")
		sb.Write(tx.Hash[:])
		sb.WriteString("blockHeight:" + strconv.Itoa(int(tx.BlockHeight)))
		sb.WriteString("index:" + strconv.Itoa(int(tx.Index)))

		writeOutPointBalance := func(ob *entity.OutPointBalance) {
			sb.WriteString("pkScript:")
			sb.Write(ob.PkScript)
			sb.WriteString("runeId:" + ob.RuneId.String())
			sb.WriteString("amount:" + ob.Amount.String())
			sb.WriteString("index:" + strconv.Itoa(int(ob.Index)))
			sb.WriteString("txHash:")
			sb.Write(ob.TxHash[:])
			sb.WriteString("txOutIndex:" + strconv.Itoa(int(ob.TxOutIndex)))
			sb.WriteString(";")
		}
		// sort inputs to ensure order
		slices.SortFunc(tx.Inputs, func(t1, t2 *entity.OutPointBalance) int {
			if t1.Index != t2.Index {
				return int(t1.Index) - int(t2.Index)
			}
			return t1.RuneId.Cmp(t2.RuneId)
		})

		sb.WriteString("in:")
		for _, in := range tx.Inputs {
			writeOutPointBalance(in)
		}
		// sort outputs to ensure order
		slices.SortFunc(tx.Inputs, func(t1, t2 *entity.OutPointBalance) int {
			if t1.Index != t2.Index {
				return int(t1.Index) - int(t2.Index)
			}
			return t1.RuneId.Cmp(t2.RuneId)
		})
		sb.WriteString("out:")
		for _, out := range tx.Outputs {
			writeOutPointBalance(out)
		}

		mintsKeys := lo.Keys(tx.Mints)
		slices.SortFunc(mintsKeys, func(t1, t2 runes.RuneId) int {
			return t1.Cmp(t2)
		})
		sb.WriteString("mints:")
		for _, runeId := range mintsKeys {
			amount := tx.Mints[runeId]
			sb.WriteString(runeId.String())
			sb.WriteString(amount.String())
			sb.WriteString(";")
		}

		burnsKeys := lo.Keys(tx.Burns)
		slices.SortFunc(mintsKeys, func(t1, t2 runes.RuneId) int {
			return t1.Cmp(t2)
		})
		sb.WriteString("burns:")
		for _, runeId := range burnsKeys {
			amount := tx.Burns[runeId]
			sb.WriteString(runeId.String())
			sb.WriteString(amount.String())
			sb.WriteString(";")
		}
		sb.WriteString("runeEtched:" + strconv.FormatBool(tx.RuneEtched))

		sb.Write(serializeRunestoneForEventHash(tx.Runestone))
		sb.WriteString(";")
	}
	return []byte(sb.String()), nil
}

func serializeRunestoneForEventHash(r *runes.Runestone) []byte {
	if r == nil {
		return []byte("rune:nil")
	}
	var sb strings.Builder
	sb.WriteString("rune:")
	if r.Etching != nil {
		etching := r.Etching
		sb.WriteString("etching:")
		if etching.Divisibility != nil {
			sb.WriteString("divisibility:" + strconv.Itoa(int(*etching.Divisibility)))
		}
		if etching.Premine != nil {
			sb.WriteString("premine:" + etching.Premine.String())
		}
		if etching.Rune != nil {
			sb.WriteString("rune:" + etching.Rune.String())
		}
		if etching.Spacers != nil {
			sb.WriteString("spacers:" + strconv.Itoa(int(*etching.Spacers)))
		}
		if etching.Symbol != nil {
			sb.WriteString("symbol:" + string(*etching.Symbol))
		}
		if etching.Terms != nil {
			terms := etching.Terms
			if terms.Amount != nil {
				sb.WriteString("amount:" + terms.Amount.String())
			}
			if terms.Cap != nil {
				sb.WriteString("cap:" + terms.Cap.String())
			}
			if terms.HeightStart != nil {
				sb.WriteString("heightStart:" + strconv.Itoa(int(*terms.HeightStart)))
			}
			if terms.HeightEnd != nil {
				sb.WriteString("heightEnd:" + strconv.Itoa(int(*terms.HeightEnd)))
			}
			if terms.OffsetStart != nil {
				sb.WriteString("offsetStart:" + strconv.Itoa(int(*terms.OffsetStart)))
			}
			if terms.OffsetEnd != nil {
				sb.WriteString("offsetEnd:" + strconv.Itoa(int(*terms.OffsetEnd)))
			}
		}
		if etching.Turbo {
			sb.WriteString("turbo:" + strconv.FormatBool(etching.Turbo))
		}
	}
	if len(r.Edicts) > 0 {
		sb.WriteString("edicts:")
		// don't sort edicts, order must be kept the same because of delta encoding
		for _, edict := range r.Edicts {
			sb.WriteString(edict.Id.String() + edict.Amount.String() + strconv.Itoa(edict.Output) + ";")
		}
	}
	if r.Mint != nil {
		sb.WriteString("mint:" + r.Mint.String())
	}
	if r.Pointer != nil {
		sb.WriteString("pointer:" + strconv.Itoa(int(*r.Pointer)))
	}
	sb.WriteString("cenotaph:" + strconv.FormatBool(r.Cenotaph))
	sb.WriteString("flaws:" + strconv.Itoa(int(r.Flaws)))
	return []byte(sb.String())
}
