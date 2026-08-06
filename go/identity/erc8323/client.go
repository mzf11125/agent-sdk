// Package erc8323 provides a Go client for ERC-8323 — IAgentSourceBinding,
// Source-Token Agent Binding for ERC-8004.
//
// ERC-8323 derives an ERC-8004 agent identity from a pre-existing ERC-721
// token in a single bound collection: registerWithSource mints a new agent to
// the caller and records (boundCollection, sourceTokenId) as immutable
// provenance via exactly one SourceNFTLinked event per agent. Live ownership
// is exposed separately and re-checked at query time by
// isSourceNFTOwnershipValid. The contract performs no verification and no
// hash derivation, so this package performs no recompute — source binding is
// an on-chain fact, read directly with view calls.
package erc8323

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

// ErrNoSigner is returned by Register when the client was constructed
// without a private key.
var ErrNoSigner = errors.New("erc8323: registerWithSource requires a signer key (NewSourceBindingClient with non-nil key)")

// SourceBindingInterfaceID is the ERC-165 interface id of IAgentSourceBinding
// (0x27eba962), the XOR of the five function selectors boundCollection ^
// registerWithSource ^ getSourceNFT ^ hasSourceNFT ^
// isSourceNFTOwnershipValid.
var SourceBindingInterfaceID = [4]byte{0x27, 0xeb, 0xa9, 0x62}

// SourceBindingClient reads source-binding state from and registers agents on
// a deployed ERC-8323 IAgentSourceBinding contract. BoundCollection,
// GetSourceNFT, HasSourceNFT, IsSourceNFTOwnershipValid, OwnerOf and
// SupportsSourceBinding are read-only view calls (no gas, no key); Register
// broadcasts a payable transaction and therefore needs a signer key — pass
// nil to NewSourceBindingClient for a read-only client.
type SourceBindingClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs registerWithSource; nil for read-only clients
}

// NewSourceBindingClient creates a client bound to a deployed ERC-8323
// IAgentSourceBinding contract. key signs the register transaction — pass nil
// for a read-only client (BoundCollection, GetSourceNFT, HasSourceNFT,
// IsSourceNFTOwnershipValid, OwnerOf, SupportsSourceBinding only).
func NewSourceBindingClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *SourceBindingClient {
	return &SourceBindingClient{rpc: rpc, address: addr, key: key}
}

// Register broadcasts registerWithSource(sourceTokenID) with the given value
// (wei), waits for the transaction to be mined, and returns the minted agent
// id parsed from the SourceNFTLinked event. The source token is checked once
// via ownerOf and never locked, escrowed, or transferred (non-custodial).
//
// registerWithSource is payable: pass the registry's mint price (wei) if it
// charges one — the call reverts on insufficient value (a real deployed
// registry, e.g. Merlini's AgentIdentityRegistry, gates on
// require(msg.value == mintPrice); the testkit mock enforces the same).
// Pass nil for value to send no msg.value. Returns ErrNoSigner if the client
// has no private key.
func (c *SourceBindingClient) Register(ctx context.Context, sourceTokenID *big.Int, value *big.Int) (*big.Int, error) {
	if c.key == nil {
		return nil, ErrNoSigner
	}
	receipt, err := c.transactAndWait(ctx, value, "registerWithSource", sourceTokenID)
	if err != nil {
		return nil, err
	}
	return parseSourceNFTLinked(receipt)
}

// BoundCollection reads the source ERC-721 collection this registry is bound
// to. Immutable for the life of the registry.
func (c *SourceBindingClient) BoundCollection(ctx context.Context) (common.Address, error) {
	out, err := c.callView(ctx, "boundCollection")
	if err != nil {
		return common.Address{}, err
	}
	vals, err := c.outputs("boundCollection", out)
	if err != nil {
		return common.Address{}, err
	}
	if len(vals) != 1 {
		return common.Address{}, fmt.Errorf("erc8323: boundCollection returned %d outputs, want 1", len(vals))
	}
	collection, ok := vals[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("erc8323: boundCollection output is %T, want common.Address", vals[0])
	}
	return collection, nil
}

// GetSourceNFT reads the immutable source token an agent was derived from.
// Errors if agentID does not exist or has no source binding.
func (c *SourceBindingClient) GetSourceNFT(ctx context.Context, agentID *big.Int) (SourceNFT, error) {
	out, err := c.callView(ctx, "getSourceNFT", agentID)
	if err != nil {
		return SourceNFT{}, err
	}
	vals, err := c.outputs("getSourceNFT", out)
	if err != nil {
		return SourceNFT{}, err
	}
	if len(vals) != 2 {
		return SourceNFT{}, fmt.Errorf("erc8323: getSourceNFT returned %d outputs, want 2", len(vals))
	}
	collection, ok := vals[0].(common.Address)
	if !ok {
		return SourceNFT{}, fmt.Errorf("erc8323: getSourceNFT sourceContract is %T, want common.Address", vals[0])
	}
	tokenID, ok := vals[1].(*big.Int)
	if !ok {
		return SourceNFT{}, fmt.Errorf("erc8323: getSourceNFT sourceTokenId is %T, want *big.Int", vals[1])
	}
	return SourceNFT{SourceContract: collection, SourceTokenID: tokenID}, nil
}

// HasSourceNFT reports whether agentID has a recorded source binding.
func (c *SourceBindingClient) HasSourceNFT(ctx context.Context, agentID *big.Int) (bool, error) {
	out, err := c.callView(ctx, "hasSourceNFT", agentID)
	if err != nil {
		return false, err
	}
	return c.singleBool("hasSourceNFT", out)
}

// IsSourceNFTOwnershipValid reports whether the source token is still under
// the control of agentID — the live 3-case check: the agent's direct owner,
// its canonical ERC-6551 token-bound account (pinned to the registry's
// declared implementation + salt), or the binding contract itself. Rechecked
// at query time, never cached.
func (c *SourceBindingClient) IsSourceNFTOwnershipValid(ctx context.Context, agentID *big.Int) (bool, error) {
	out, err := c.callView(ctx, "isSourceNFTOwnershipValid", agentID)
	if err != nil {
		return false, err
	}
	return c.singleBool("isSourceNFTOwnershipValid", out)
}

// OwnerOf reads the ERC-721 owner of an agent id (a compliant registry MUST
// also implement ERC-721). Errors for non-existent agents.
func (c *SourceBindingClient) OwnerOf(ctx context.Context, agentID *big.Int) (common.Address, error) {
	out, err := c.callView(ctx, "ownerOf", agentID)
	if err != nil {
		return common.Address{}, err
	}
	vals, err := c.outputs("ownerOf", out)
	if err != nil {
		return common.Address{}, err
	}
	if len(vals) != 1 {
		return common.Address{}, fmt.Errorf("erc8323: ownerOf returned %d outputs, want 1", len(vals))
	}
	owner, ok := vals[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("erc8323: ownerOf output is %T, want common.Address", vals[0])
	}
	return owner, nil
}

// SupportsSourceBinding reports whether the contract advertises
// IAgentSourceBinding via ERC-165 (interface id 0x27eba962).
func (c *SourceBindingClient) SupportsSourceBinding(ctx context.Context) (bool, error) {
	out, err := c.callView(ctx, "supportsInterface", SourceBindingInterfaceID)
	if err != nil {
		return false, err
	}
	return c.singleBool("supportsInterface", out)
}

// callView packs the function inputs and performs a read-only eth_call
// against the bound contract address.
func (c *SourceBindingClient) callView(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	a, err := SourceBindingABI()
	if err != nil {
		return nil, fmt.Errorf("erc8323: parse ABI: %w", err)
	}
	// a.Pack prepends the 4-byte method selector to the packed arguments.
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8323: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &c.address, Data: data}
	return c.rpc.CallContract(ctx, msg, nil)
}

// outputs unpacks the raw call result with the method's declared output types.
func (c *SourceBindingClient) outputs(methodName string, data []byte) ([]interface{}, error) {
	a, err := SourceBindingABI()
	if err != nil {
		return nil, fmt.Errorf("erc8323: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8323: ABI has no method %q", methodName)
	}
	vals, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("erc8323: unpack %s outputs: %w", methodName, err)
	}
	return vals, nil
}

// singleBool decodes a single-bool view result.
func (c *SourceBindingClient) singleBool(methodName string, data []byte) (bool, error) {
	vals, err := c.outputs(methodName, data)
	if err != nil {
		return false, err
	}
	if len(vals) != 1 {
		return false, fmt.Errorf("erc8323: %s returned %d outputs, want 1", methodName, len(vals))
	}
	v, ok := vals[0].(bool)
	if !ok {
		return false, fmt.Errorf("erc8323: %s output is %T, want bool", methodName, vals[0])
	}
	return v, nil
}

// transactAndWait packs the method inputs via the ABI (a.Pack prepends the
// 4-byte method selector), sets the transaction value (msg.value for the
// payable registerWithSource), signs, estimates gas and broadcasts through
// the same rpc client, then waits for the receipt — erroring if the
// transaction reverted. Gas limit, base fee and nonce are resolved against
// the live node; the chain id is fetched from the RPC at call time.
func (c *SourceBindingClient) transactAndWait(ctx context.Context, value *big.Int, methodName string, args ...interface{}) (*types.Receipt, error) {
	a, err := SourceBindingABI()
	if err != nil {
		return nil, fmt.Errorf("erc8323: parse ABI: %w", err)
	}
	chainID, err := c.rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erc8323: fetch chain id: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.key, chainID)
	if err != nil {
		return nil, fmt.Errorf("erc8323: create transactor: %w", err)
	}
	if value != nil {
		auth.Value = value
	}
	bound := bind.NewBoundContract(c.address, a, c.rpc, c.rpc, c.rpc)
	tx, err := bound.Transact(auth, methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8323: %s: %w", methodName, err)
	}
	receipt, err := bind.WaitMined(ctx, c.rpc, tx)
	if err != nil {
		return nil, fmt.Errorf("erc8323: wait for %s to mine: %w", methodName, err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("erc8323: %s reverted (tx %s)", methodName, tx.Hash().Hex())
	}
	return receipt, nil
}

// parseSourceNFTLinked extracts the minted agent id from the SourceNFTLinked
// event log. agentId (indexed uint256, topic 1) and sourceContract (indexed
// address, topic 2) live in the topics; sourceTokenId (uint256) is in the log
// data.
func parseSourceNFTLinked(receipt *types.Receipt) (*big.Int, error) {
	a, err := SourceBindingABI()
	if err != nil {
		return nil, fmt.Errorf("erc8323: parse ABI: %w", err)
	}
	evt, ok := a.Events["SourceNFTLinked"]
	if !ok {
		return nil, errors.New("erc8323: ABI has no SourceNFTLinked event")
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] != evt.ID {
			continue
		}
		if len(log.Topics) != 3 {
			return nil, fmt.Errorf("erc8323: SourceNFTLinked log has %d topics, want 3", len(log.Topics))
		}
		return new(big.Int).SetBytes(log.Topics[1].Bytes()), nil
	}
	return nil, errors.New("erc8323: SourceNFTLinked event not found in receipt")
}
