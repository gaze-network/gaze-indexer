package brc20

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/samber/lo"
)

const eventHashSeparator = "|"

func getEventDeployString(event *entity.EventDeploy) string {
	var sb strings.Builder
	sb.WriteString("deploy-inscribe;")
	sb.WriteString(event.InscriptionId.String() + ";")
	sb.WriteString(hex.EncodeToString(event.PkScript) + ";")
	sb.WriteString(event.Tick + ";")
	sb.WriteString(event.OriginalTick + ";")
	sb.WriteString(event.TotalSupply.StringFixed(int32(event.Decimals)) + ";")
	sb.WriteString(strconv.Itoa(int(event.Decimals)) + ";")
	sb.WriteString(event.LimitPerMint.StringFixed(int32(event.Decimals)) + ";")
	sb.WriteString(lo.Ternary(event.IsSelfMint, "True", "False"))
	return sb.String()
}

func getEventMintString(event *entity.EventMint, decimals uint16) string {
	var sb strings.Builder
	var parentId string
	if event.ParentId != nil {
		parentId = event.ParentId.String()
	}
	sb.WriteString("mint-inscribe;")
	sb.WriteString(event.InscriptionId.String() + ";")
	sb.WriteString(hex.EncodeToString(event.PkScript) + ";")
	sb.WriteString(event.Tick + ";")
	sb.WriteString(event.OriginalTick + ";")
	sb.WriteString(event.Amount.StringFixed(int32(decimals)) + ";")
	sb.WriteString(parentId)
	return sb.String()
}

func getEventInscribeTransferString(event *entity.EventInscribeTransfer, decimals uint16) string {
	var sb strings.Builder
	sb.WriteString("inscribe-transfer;")
	sb.WriteString(event.InscriptionId.String() + ";")
	sb.WriteString(hex.EncodeToString(event.PkScript) + ";")
	sb.WriteString(event.Tick + ";")
	sb.WriteString(event.OriginalTick + ";")
	sb.WriteString(event.Amount.StringFixed(int32(decimals)))
	return sb.String()
}

func getEventTransferTransferString(event *entity.EventTransferTransfer, decimals uint16) string {
	var sb strings.Builder
	sb.WriteString("transfer-transfer;")
	sb.WriteString(event.InscriptionId.String() + ";")
	sb.WriteString(hex.EncodeToString(event.FromPkScript) + ";")
	if event.SpentAsFee {
		sb.WriteString(";")
	} else {
		sb.WriteString(hex.EncodeToString(event.ToPkScript) + ";")
	}
	sb.WriteString(event.Tick + ";")
	sb.WriteString(event.OriginalTick + ";")
	sb.WriteString(event.Amount.StringFixed(int32(decimals)))
	return sb.String()
}
