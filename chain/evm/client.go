package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/renproject/pack"
	"github.com/renprotocol/multichain/api/account"
	"github.com/renprotocol/multichain/api/address"
	"github.com/renprotocol/multichain/api/contract"
)

const (
	// DefaultClientRPCURL is the RPC URL used by default, to interact with the
	// ethereum node.
	DefaultClientRPCURL = "http://127.0.0.1:8545/"
)

// Client holds the underlying RPC client instance.
type Client struct {
	EthClient *ethclient.Client
	ChainID   *big.Int
}

// NewClient creates and returns a new JSON-RPC client to the Ethereum node
func NewClient(rpcURL string, chainID *big.Int) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dialing url: %v", rpcURL)
	}
	clientChainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("mismatched chain id: expected %v, got %v", chainID, clientChainID)
	}
	return &Client{
		client,
		chainID,
	}, nil
}

// LatestBlock returns the block number at the current chain head.
func (client *Client) LatestBlock(ctx context.Context) (pack.U64, error) {
	header, err := client.EthClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return pack.NewU64(0), fmt.Errorf("fetching header: %v", err)
	}
	return pack.NewU64(header.Number.Uint64()), nil
}

// Tx returns the transaction uniquely identified by the given transaction
// hash. It also returns the number of confirmations for the transaction.
func (client *Client) Tx(ctx context.Context, txID pack.Bytes) (account.Tx, pack.U64, error) {
	tx, pending, err := client.EthClient.TransactionByHash(ctx, common.BytesToHash(txID))
	if err != nil {
		return nil, pack.NewU64(0), fmt.Errorf(fmt.Sprintf("fetching tx by hash '%v': %v", txID, err))
	}

	// Check the chain id for replay-protected tx.
	if tx.Protected() {
		chainID := tx.ChainId()
		if client.ChainID != nil {
			if chainID == nil {
				return nil, 0, fmt.Errorf("nil chain ID")
			}
			if chainID.Cmp(client.ChainID) != 0 {
				return nil, 0, fmt.Errorf("invalid chain ID, expected = %v, got = %v", client.ChainID.String(), chainID.String())
			}
		}
	}

	// If the transaction is still pending, use default EIP-155 Signer.
	pendingTx := Tx{
		EthTx:  tx,
		Signer: types.NewEIP155Signer(client.ChainID),
	}
	if pending {
		// Transaction has not been included in a block yet.
		return nil, 0, fmt.Errorf("tx %v is pending", txID)
	}

	receipt, err := client.EthClient.TransactionReceipt(ctx, common.BytesToHash(txID))
	if err != nil {
		return nil, pack.NewU64(0), fmt.Errorf("fetching recipt for tx %v : %v", txID, err)
	}

	if receipt == nil {
		// Transaction has 0 confirmations.
		return &pendingTx, 0, nil
	}

	if receipt.Status == 0 {
		// Transaction has been reverted.
		return nil, pack.NewU64(0), fmt.Errorf("tx %v reverted, reciept status 0", txID)
	}

	// Transaction has been confirmed.
	confirmedTx := Tx{
		tx,
		types.LatestSignerForChainID(client.ChainID),
	}

	header, err := client.EthClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, pack.NewU64(0), fmt.Errorf("fetching header : %v", err)
	}

	return &confirmedTx, pack.NewU64(header.Number.Uint64() - receipt.BlockNumber.Uint64()), nil
}

// SubmitTx to the underlying blockchain network.
func (client *Client) SubmitTx(ctx context.Context, tx account.Tx) error {
	switch tx := tx.(type) {
	case *Tx:
		err := client.EthClient.SendTransaction(ctx, tx.EthTx)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("sending transaction '%v': %v", tx.Hash(), err))
		}
		return nil
	default:
		return fmt.Errorf("expected type %T, got type %T", new(Tx), tx)
	}
}

// AccountNonce returns the current nonce of the account. This is the nonce to
// be used while building a new transaction.
func (client *Client) AccountNonce(ctx context.Context, addr address.Address) (pack.U256, error) {
	targetAddr, err := NewAddressFromHex(string(pack.String(addr)))
	if err != nil {
		return pack.U256{}, fmt.Errorf("bad to address '%v': %v", addr, err)
	}
	nonce, err := client.EthClient.NonceAt(ctx, common.Address(targetAddr), nil)
	if err != nil {
		return pack.U256{}, fmt.Errorf("failed to get nonce for '%v': %v", addr, err)
	}

	return pack.NewU256FromU64(pack.NewU64(nonce)), nil
}

// AccountBalance returns the account balancee for a given address.
func (client *Client) AccountBalance(ctx context.Context, addr address.Address) (pack.U256, error) {
	targetAddr, err := NewAddressFromHex(string(pack.String(addr)))
	if err != nil {
		return pack.U256{}, fmt.Errorf("bad to address '%v': %v", addr, err)
	}
	balance, err := client.EthClient.BalanceAt(ctx, common.Address(targetAddr), nil)
	if err != nil {
		return pack.U256{}, fmt.Errorf("failed to get balance for '%v': %v", addr, err)
	}

	return pack.NewU256FromInt(balance), nil
}

// CallContract implements the multichain Contract API.
func (client *Client) CallContract(ctx context.Context, program address.Address, calldata contract.CallData) (pack.Bytes, error) {
	targetAddr, err := NewAddressFromHex(string(pack.String(program)))
	if err != nil {
		return nil, fmt.Errorf("bad to address '%v': %v", program, err)
	}
	addr := common.Address(targetAddr)

	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: calldata,
	}
	return client.EthClient.CallContract(ctx, callMsg, nil)
}
