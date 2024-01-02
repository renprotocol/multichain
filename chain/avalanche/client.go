package avalanche

import (
	"github.com/renprotocol/multichain/chain/evm"
)

const (
	// DefaultClientRPCURL is the RPC URL used by default, to interact with the
	// avalanche node.
	DefaultClientRPCURL = "http://127.0.0.1:9650/ext/bc/C/rpc"
)

// Client re-exports evm.Client.
type Client = evm.Client

// NewClient re-exports evm.NewClient.
var NewClient = evm.NewClient
