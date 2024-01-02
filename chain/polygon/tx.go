package polygon

import (
	"github.com/renprotocol/multichain/chain/evm"
)

type (
	// TxBuilder re-exports evm.TxBuilder.
	TxBuilder = evm.TxBuilder

	// Tx re-exports evm.Tx.
	Tx = evm.Tx
)

// NewTxBuilder re-exports evm.NewTxBuilder.
var NewTxBuilder = evm.NewTxBuilder
