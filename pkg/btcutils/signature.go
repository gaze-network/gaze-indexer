package btcutils

import (
	"github.com/Cleverse/go-utilities/utils"
	verifier "github.com/bitonicnl/verify-signed-message/pkg"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cockroachdb/errors"
)

func VerifySignature(address string, message string, sigBase64 string, defaultNet ...*chaincfg.Params) error {
	net := utils.DefaultOptional(defaultNet, &chaincfg.MainNetParams)
	_, err := verifier.VerifyWithChain(verifier.SignedMessage{
		Address:   address,
		Message:   message,
		Signature: sigBase64,
	}, net)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
