package btcutils

import (
	"testing"

	"github.com/gaze-network/indexer-network/common"
	"github.com/stretchr/testify/assert"
)

func TestVerifySignature(t *testing.T) {
	{
		message := "Test123"
		address := "18J72YSM9pKLvyXX1XAjFXA98zeEvxBYmw"
		signature := "Gzhfsw0ItSrrTCChykFhPujeTyAcvVxiXwywxpHmkwFiKuUR2ETbaoFcocmcSshrtdIjfm8oXlJoTOLosZp3Yc8="
		network := common.NetworkMainnet

		err := VerifySignature(address, message, signature, network)
		assert.NoError(t, err)
	}
	{
		address := "tb1qr97cuq4kvq7plfetmxnl6kls46xaka78n2288z"
		message := "The outage comes at a time when bitcoin has been fast approaching new highs not seen since June 26, 2019."
		signature := "H/bSByRH7BW1YydfZlEx9x/nt4EAx/4A691CFlK1URbPEU5tJnTIu4emuzkgZFwC0ptvKuCnyBThnyLDCqPqT10="
		network := common.NetworkTestnet

		err := VerifySignature(address, message, signature, network)
		assert.NoError(t, err)
	}
	{
		// Missmatch address
		address := "tb1qp7y2ywgrv8a4t9h47yphtgj8w759rk6vgd9ran"
		message := "The outage comes at a time when bitcoin has been fast approaching new highs not seen since June 26, 2019."
		signature := "H/bSByRH7BW1YydfZlEx9x/nt4EAx/4A691CFlK1URbPEU5tJnTIu4emuzkgZFwC0ptvKuCnyBThnyLDCqPqT10="
		network := common.NetworkTestnet

		err := VerifySignature(address, message, signature, network)
		assert.Error(t, err)
	}
	{
		// Missmatch signature
		address := "tb1qr97cuq4kvq7plfetmxnl6kls46xaka78n2288z"
		message := "The outage comes at a time when bitcoin has been fast approaching new highs not seen since June 26, 2019."
		signature := "Gzhfsw0ItSrrTCChykFhPujeTyAcvVxiXwywxpHmkwFiKuUR2ETbaoFcocmcSshrtdIjfm8oXlJoTOLosZp3Yc8="
		network := common.NetworkTestnet

		err := VerifySignature(address, message, signature, network)
		assert.Error(t, err)
	}
	{
		// Missmatch message
		address := "tb1qr97cuq4kvq7plfetmxnl6kls46xaka78n2288z"
		message := "Hello World"
		signature := "H/bSByRH7BW1YydfZlEx9x/nt4EAx/4A691CFlK1URbPEU5tJnTIu4emuzkgZFwC0ptvKuCnyBThnyLDCqPqT10="
		network := common.NetworkTestnet

		err := VerifySignature(address, message, signature, network)
		assert.Error(t, err)
	}
}
