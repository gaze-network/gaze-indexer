package nodesale

import "github.com/btcsuite/btcd/txscript"

func extractTapScript(witness [][]byte) (txscript.ScriptTokenizer, bool) {
	witness = removeAnnexFromWitness(witness)
	if len(witness) < 2 {
		return txscript.ScriptTokenizer{}, false
	}
	script := witness[len(witness)-2]

	return txscript.MakeScriptTokenizer(0, script), true
}

func removeAnnexFromWitness(witness [][]byte) [][]byte {
	if len(witness) >= 2 && len(witness[len(witness)-1]) > 0 && witness[len(witness)-1][0] == txscript.TaprootAnnexTag {
		return witness[:len(witness)-1]
	}
	return witness
}
