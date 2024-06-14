package btcutils

import (
	verifier "github.com/bitonicnl/verify-signed-message/pkg"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
)

func VerifySignature(address string, message string, sigBase64 string, network common.Network) error {
	_, err := verifier.VerifyWithChain(verifier.SignedMessage{
		Address:   address,
		Message:   message,
		Signature: sigBase64,
	}, network.ChainParams())
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
