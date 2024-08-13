package nodesale

import "github.com/btcsuite/btcd/txscript"

func extractTapScript(witness [][]byte) (tokenizer txscript.ScriptTokenizer, controlBlock *txscript.ControlBlock, isTapScript bool) {
	witness = removeAnnexFromWitness(witness)
	if len(witness) < 2 {
		return txscript.ScriptTokenizer{}, nil, false
	}
	script := witness[len(witness)-2]
	rawControl := witness[len(witness)-1]
	parsedControl, err := txscript.ParseControlBlock(rawControl)
	if err != nil {
		return txscript.ScriptTokenizer{}, nil, false
	}

	return txscript.MakeScriptTokenizer(0, script), parsedControl, true
}

func removeAnnexFromWitness(witness [][]byte) [][]byte {
	if len(witness) >= 2 && len(witness[len(witness)-1]) > 0 && witness[len(witness)-1][0] == txscript.TaprootAnnexTag {
		return witness[:len(witness)-1]
	}
	return witness
}
