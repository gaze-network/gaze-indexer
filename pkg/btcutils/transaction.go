package btcutils

const (
	// TxVersion is the current latest supported transaction version.
	TxVersion = 2

	// MaxTxInSequenceNum is the maximum sequence number the sequence field
	// of a transaction input can be.
	MaxTxInSequenceNum uint32 = 0xffffffff
)
