package btcutils

import (
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifySignature(t *testing.T) {
	{
		message := "Test123"
		address := "18J72YSM9pKLvyXX1XAjFXA98zeEvxBYmw"
		signature := "Gzhfsw0ItSrrTCChykFhPujeTyAcvVxiXwywxpHmkwFiKuUR2ETbaoFcocmcSshrtdIjfm8oXlJoTOLosZp3Yc8="
		network := &chaincfg.MainNetParams

		err := VerifySignature(address, message, signature, network)
		assert.NoError(t, err)
	}
	{
		address := "tb1qr97cuq4kvq7plfetmxnl6kls46xaka78n2288z"
		message := "The outage comes at a time when bitcoin has been fast approaching new highs not seen since June 26, 2019."
		signature := "H/bSByRH7BW1YydfZlEx9x/nt4EAx/4A691CFlK1URbPEU5tJnTIu4emuzkgZFwC0ptvKuCnyBThnyLDCqPqT10="
		network := &chaincfg.TestNet3Params

		err := VerifySignature(address, message, signature, network)
		assert.NoError(t, err)
	}
	{
		// Missmatch address
		address := "tb1qp7y2ywgrv8a4t9h47yphtgj8w759rk6vgd9ran"
		message := "The outage comes at a time when bitcoin has been fast approaching new highs not seen since June 26, 2019."
		signature := "H/bSByRH7BW1YydfZlEx9x/nt4EAx/4A691CFlK1URbPEU5tJnTIu4emuzkgZFwC0ptvKuCnyBThnyLDCqPqT10="
		network := &chaincfg.TestNet3Params

		err := VerifySignature(address, message, signature, network)
		assert.Error(t, err)
	}
	{
		// Missmatch signature
		address := "tb1qr97cuq4kvq7plfetmxnl6kls46xaka78n2288z"
		message := "The outage comes at a time when bitcoin has been fast approaching new highs not seen since June 26, 2019."
		signature := "Gzhfsw0ItSrrTCChykFhPujeTyAcvVxiXwywxpHmkwFiKuUR2ETbaoFcocmcSshrtdIjfm8oXlJoTOLosZp3Yc8="
		network := &chaincfg.TestNet3Params

		err := VerifySignature(address, message, signature, network)
		assert.Error(t, err)
	}
	{
		// Missmatch message
		address := "tb1qr97cuq4kvq7plfetmxnl6kls46xaka78n2288z"
		message := "Hello World"
		signature := "H/bSByRH7BW1YydfZlEx9x/nt4EAx/4A691CFlK1URbPEU5tJnTIu4emuzkgZFwC0ptvKuCnyBThnyLDCqPqT10="
		network := &chaincfg.TestNet3Params

		err := VerifySignature(address, message, signature, network)
		assert.Error(t, err)
	}
	{
		// Missmatch network
		address := "tb1qr97cuq4kvq7plfetmxnl6kls46xaka78n2288z"
		message := "The outage comes at a time when bitcoin has been fast approaching new highs not seen since June 26, 2019."
		signature := "H/bSByRH7BW1YydfZlEx9x/nt4EAx/4A691CFlK1URbPEU5tJnTIu4emuzkgZFwC0ptvKuCnyBThnyLDCqPqT10="
		network := &chaincfg.MainNetParams

		err := VerifySignature(address, message, signature, network)
		assert.Error(t, err)
	}
}

func TestSignTxInput(t *testing.T) {
	generateTxAndPrevTxOutFromPkScript := func(pkScript []byte) (*wire.MsgTx, *wire.TxOut) {
		tx := wire.NewMsgTx(wire.TxVersion)
		tx.AddTxIn(&wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Index: 1,
			},
		})
		txOut := &wire.TxOut{
			Value: 1e8, PkScript: pkScript,
		}
		tx.AddTxOut(txOut)
		// using same value and pkScript as input for simplicity
		return tx, txOut
	}
	verifySignedTx := func(t *testing.T, signedTx *wire.MsgTx, prevTxOut *wire.TxOut) {
		t.Helper()

		prevOutFetcher := txscript.NewCannedPrevOutputFetcher(prevTxOut.PkScript, prevTxOut.Value)
		sigHashes := txscript.NewTxSigHashes(signedTx, prevOutFetcher)
		vm, err := txscript.NewEngine(
			prevTxOut.PkScript, signedTx, 0, txscript.StandardVerifyFlags,
			nil, sigHashes, prevTxOut.Value, prevOutFetcher,
		)
		require.NoError(t, err)
		require.NoError(t, vm.Execute(), "error during signature verification") // no error means success
	}

	privKey, _ := btcec.NewPrivateKey()
	t.Run("P2TR input", func(t *testing.T) {
		taprootKey := txscript.ComputeTaprootKeyNoScript(privKey.PubKey())

		pkScript, err := txscript.PayToTaprootScript(taprootKey)
		require.NoError(t, err)

		tx, prevTxOut := generateTxAndPrevTxOutFromPkScript(pkScript)
		signedTx, err := SignTxInput(
			tx, privKey, prevTxOut, 0,
		)
		require.NoError(t, err)
		verifySignedTx(t, signedTx, prevTxOut)
	})
	t.Run("tapscript input", func(t *testing.T) {
		internalKey := privKey.PubKey()

		// Our script will be a simple OP_CHECKSIG as the sole leaf of a
		// tapscript tree.
		builder := txscript.NewScriptBuilder()
		builder.AddData(schnorr.SerializePubKey(internalKey))
		builder.AddOp(txscript.OP_CHECKSIG)
		tapScript, err := builder.Script()
		require.NoError(t, err)

		tapLeaf := txscript.NewBaseTapLeaf(tapScript)
		tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)

		controlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(
			internalKey,
		)
		controlBlockBytes, err := controlBlock.ToBytes()
		require.NoError(t, err)

		tapScriptRootHash := tapScriptTree.RootNode.TapHash()
		outputKey := txscript.ComputeTaprootOutputKey(
			internalKey, tapScriptRootHash[:],
		)
		p2trScript, err := txscript.PayToTaprootScript(outputKey)
		require.NoError(t, err)

		tx, prevTxOut := generateTxAndPrevTxOutFromPkScript(p2trScript)
		tx.TxIn[0].Witness = wire.TxWitness{
			{},
			tapScript,
			controlBlockBytes,
		}
		signedTx, err := SignTxInputTapScript(
			tx, privKey, prevTxOut, 0,
		)
		require.NoError(t, err)
		verifySignedTx(t, signedTx, prevTxOut)
	})
	t.Run("P2WPKH input", func(t *testing.T) {
		pubKey := privKey.PubKey()
		pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())

		pkScript, err := txscript.NewScriptBuilder().
			AddOp(txscript.OP_0).
			AddData(pubKeyHash).
			Script()

		tx, prevTxOut := generateTxAndPrevTxOutFromPkScript(pkScript)
		signedTx, err := SignTxInput(
			tx, privKey, prevTxOut, 0,
		)
		require.NoError(t, err)
		verifySignedTx(t, signedTx, prevTxOut)
	})
	t.Run("P2PKH input", func(t *testing.T) {
		pubKey := privKey.PubKey()
		pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
		address, err := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
		pkScript, err := txscript.PayToAddrScript(address)

		tx, prevTxOut := generateTxAndPrevTxOutFromPkScript(pkScript)
		signedTx, err := SignTxInput(
			tx, privKey, prevTxOut, 0,
		)
		require.NoError(t, err)
		verifySignedTx(t, signedTx, prevTxOut)
	})
}
