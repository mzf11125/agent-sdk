// Package erc8263 provides a Go client for ERC-8263 — IOnChainProof, the
// OnChain Proof Anchor.
//
// ERC-8263 is a write-side anchor floor: an agent (or its operator) anchors
// a non-zero 32-byte commitment (proofHash) together with an identity-scheme
// byte and a 32-byte agent identifier, producing a verifiable, immutable
// timeline of agent activity via exactly one canonical AnchorProof event per
// anchor. The contract performs no verification and no profile detection, so
// this package performs no recompute — verification of proofHash belongs to
// higher-layer verifier profiles.
package erc8263

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

// ErrNoSigner is returned by Anchor and AnchorWithAux when the client was
// constructed without a private key.
var ErrNoSigner = errors.New("erc8263: Anchor requires a signer key (NewOnChainProofClient with non-nil key)")

// OnChainProofClient anchors proof hashes on-chain via
// IOnChainProof.anchor / anchorWithAux (ERC-8263) and reads back the
// AnchorProof event log — the ledger of the protocol, since the interface
// exposes no getter.
type OnChainProofClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs anchor transactions; nil for read-only clients
}

// NewOnChainProofClient creates a client bound to a deployed ERC-8263
// contract. key signs the anchor transactions — pass nil for a read-only
// client (IsAnchored only).
func NewOnChainProofClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *OnChainProofClient {
	return &OnChainProofClient{rpc: rpc, address: addr, key: key}
}

// Anchor sends IOnChainProof.anchor with empty aux, emitting the
// AnchorProof(agentIdScheme, agentId, proofHash, operator) event. The
// returned transaction is signed and broadcast; the caller can wait for its
// receipt to extract the chain/block/log position for the proof envelope.
//
// agentIdScheme is the identity scheme byte: 0x00 ANONYMOUS (agentId MUST be
// zero), 0x01 REGISTRY (non-zero registry record id, e.g. ERC-8004), 0x02
// URI_HASH (non-zero keccak256 of the canonical agent URI); 0x03+ is
// reserved. proofHash MUST be non-zero. The contract enforces these
// canonical-form guards — a violating call reverts, surfaced here as an
// error from gas estimation.
//
// Gas limit, base fee and nonce are resolved against the live node; the
// chain id is fetched from the RPC at call time. Returns ErrNoSigner if the
// client has no private key.
func (c *OnChainProofClient) Anchor(agentIdScheme uint8, agentId common.Hash, proofHash common.Hash) (*types.Transaction, error) {
	return c.transact("anchor", agentIdScheme, agentId, proofHash)
}

// AnchorWithAux sends IOnChainProof.anchorWithAux with opaque extension
// bytes for adjacent protocols (e.g. OCP digest commitments, session ids,
// parent-proof references). Same semantics and guards as Anchor; aux is
// explicitly non-normative.
func (c *OnChainProofClient) AnchorWithAux(agentIdScheme uint8, agentId common.Hash, proofHash common.Hash, aux []byte) (*types.Transaction, error) {
	return c.transact("anchorWithAux", agentIdScheme, agentId, proofHash, aux)
}

// transact packs the inputs via a.Pack(name, args...) — which prepends the
// 4-byte method selector — signs, estimates gas (GasLimit 0) and broadcasts
// through the same rpc client. Guard-violating input reverts during
// estimation and surfaces as an error here, before anything is broadcast.
func (c *OnChainProofClient) transact(method string, args ...interface{}) (*types.Transaction, error) {
	if c.key == nil {
		return nil, ErrNoSigner
	}
	a, err := OnChainProofABI()
	if err != nil {
		return nil, fmt.Errorf("erc8263: parse ABI: %w", err)
	}
	ctx := context.Background()
	chainID, err := c.rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erc8263: fetch chain id: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.key, chainID)
	if err != nil {
		return nil, fmt.Errorf("erc8263: create transactor: %w", err)
	}
	bound := bind.NewBoundContract(c.address, a, c.rpc, c.rpc, c.rpc)
	tx, err := bound.Transact(auth, method, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8263: %s: %w", method, err)
	}
	return tx, nil
}

// IsAnchored reports whether the contract has ever emitted an AnchorProof
// event for proofHash. The event log is the ledger — ERC-8263 exposes no
// on-chain getter — so this scans the contract's AnchorProof events
// (topics[2] == proofHash) from block 0.
func (c *OnChainProofClient) IsAnchored(proofHash common.Hash) (bool, error) {
	a, err := OnChainProofABI()
	if err != nil {
		return false, fmt.Errorf("erc8263: parse ABI: %w", err)
	}
	event, ok := a.Events["AnchorProof"]
	if !ok {
		return false, errors.New("erc8263: ABI has no AnchorProof event")
	}
	// Each inner slice is an OR-set for one topic position; topics[0] must
	// be the AnchorProof signature and topics[2] the proofHash (the agentId
	// and operator positions are left open).
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		Addresses: []common.Address{c.address},
		Topics:    [][]common.Hash{{event.ID}, nil, {proofHash}},
	}
	logs, err := c.rpc.FilterLogs(context.Background(), query)
	if err != nil {
		return false, fmt.Errorf("erc8263: filter AnchorProof logs: %w", err)
	}
	return len(logs) > 0, nil
}
