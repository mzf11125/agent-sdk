package erc8281

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ErrNoSigner is returned by Record when the client was constructed without
// a private key.
var ErrNoSigner = errors.New("erc8281: Record requires a signer key (NewObservationCommitmentClient with non-nil key)")

// ObservationCommitmentClient anchors commitment digests on-chain via
// IObservationCommitment.record (ERC-8281) and reads back the Recorded
// event log — the ledger of the protocol, since the interface exposes no
// getter.
type ObservationCommitmentClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs record() transactions; nil for read-only clients
}

// NewObservationCommitmentClient creates a client bound to a deployed
// ERC-8281 contract. key signs the record() transaction — pass nil for a
// read-only client (CheckRecorded only).
func NewObservationCommitmentClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *ObservationCommitmentClient {
	return &ObservationCommitmentClient{rpc: rpc, address: addr, key: key}
}

// Record commits a digest on-chain via IObservationCommitment.record,
// emitting the Recorded(digest, committer) event. The transaction is mined
// before this call returns so a subsequent CheckRecorded sees the record.
//
// Gas limit, base fee and nonce are resolved against the live node; the
// chain id is fetched from the RPC at call time. Returns ErrNoSigner if the
// client has no private key.
func (c *ObservationCommitmentClient) Record(digest common.Hash) (*types.Transaction, error) {
	if c.key == nil {
		return nil, ErrNoSigner
	}
	a, err := ObservationCommitmentABI()
	if err != nil {
		return nil, fmt.Errorf("erc8281: parse ABI: %w", err)
	}
	ctx := context.Background()
	chainID, err := c.rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erc8281: fetch chain id: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.key, chainID)
	if err != nil {
		return nil, fmt.Errorf("erc8281: create transactor: %w", err)
	}
	// Transact packs the inputs via a.Pack(name, args...) — which prepends
	// the 4-byte method selector — signs, estimates gas (GasLimit 0) and
	// broadcasts through the same rpc client.
	bound := bind.NewBoundContract(c.address, a, c.rpc, c.rpc, c.rpc)
	tx, err := bound.Transact(auth, "record", digest)
	if err != nil {
		return nil, fmt.Errorf("erc8281: record(%s): %w", digest.Hex(), err)
	}
	if _, err := bind.WaitMined(ctx, c.rpc, tx); err != nil {
		return nil, fmt.Errorf("erc8281: wait for record to mine: %w", err)
	}
	return tx, nil
}

// CheckRecorded reports whether the contract has ever emitted a Recorded
// event for digest. The event log is the ledger — ERC-8281 exposes no
// on-chain getter — so this scans the contract's Recorded events (topics[1]
// == digest) from block 0.
func (c *ObservationCommitmentClient) CheckRecorded(digest common.Hash) (bool, error) {
	a, err := ObservationCommitmentABI()
	if err != nil {
		return false, fmt.Errorf("erc8281: parse ABI: %w", err)
	}
	event, ok := a.Events["Recorded"]
	if !ok {
		return false, errors.New("erc8281: ABI has no Recorded event")
	}
	// Each inner slice is an OR-set for one topic position; topics[0] must
	// be the Recorded signature and topics[1] the digest.
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		Addresses: []common.Address{c.address},
		Topics:    [][]common.Hash{{event.ID}, {digest}},
	}
	logs, err := c.rpc.FilterLogs(context.Background(), query)
	if err != nil {
		return false, fmt.Errorf("erc8281: filter Recorded logs: %w", err)
	}
	return len(logs) > 0, nil
}
