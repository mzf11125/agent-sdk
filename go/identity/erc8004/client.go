package erc8004

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// IdentityRegistryClient reads identity state from a deployed ERC-8004
// contract. All calls are read-only view functions — no gas or broadcast
// needed.
type IdentityRegistryClient struct {
	rpc     *ethclient.Client
	address common.Address
}

// NewIdentityRegistryClient creates a client bound to a deployed ERC-8004
// contract.
func NewIdentityRegistryClient(rpc *ethclient.Client, addr common.Address) *IdentityRegistryClient {
	return &IdentityRegistryClient{rpc: rpc, address: addr}
}

// GetAgentURI reads the registration-file URI for an agent (ERC-721
// tokenURI). Returns "" for agents registered without a URI.
func (c *IdentityRegistryClient) GetAgentURI(ctx context.Context, tokenID *big.Int) (string, error) {
	out, err := c.callView(ctx, "tokenURI", tokenID)
	if err != nil {
		return "", err
	}
	vals, err := c.outputs("tokenURI", out)
	if err != nil {
		return "", err
	}
	if len(vals) != 1 {
		return "", fmt.Errorf("erc8004: tokenURI returned %d outputs, want 1", len(vals))
	}
	uri, ok := vals[0].(string)
	if !ok {
		return "", fmt.Errorf("erc8004: tokenURI output is %T, want string", vals[0])
	}
	return uri, nil
}

// GetMetadata reads an on-chain metadata value for an agent. Returns the
// empty bytes for keys that were never set.
func (c *IdentityRegistryClient) GetMetadata(ctx context.Context, agentID *big.Int, metadataKey string) ([]byte, error) {
	out, err := c.callView(ctx, "getMetadata", agentID, metadataKey)
	if err != nil {
		return nil, err
	}
	vals, err := c.outputs("getMetadata", out)
	if err != nil {
		return nil, err
	}
	if len(vals) != 1 {
		return nil, fmt.Errorf("erc8004: getMetadata returned %d outputs, want 1", len(vals))
	}
	value, ok := vals[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("erc8004: getMetadata output is %T, want []byte", vals[0])
	}
	return value, nil
}

// GetAgentWallet reads the agent's current payment wallet. Returns the zero
// address when no wallet was set.
func (c *IdentityRegistryClient) GetAgentWallet(ctx context.Context, agentID *big.Int) (common.Address, error) {
	out, err := c.callView(ctx, "getAgentWallet", agentID)
	if err != nil {
		return common.Address{}, err
	}
	vals, err := c.outputs("getAgentWallet", out)
	if err != nil {
		return common.Address{}, err
	}
	if len(vals) != 1 {
		return common.Address{}, fmt.Errorf("erc8004: getAgentWallet returned %d outputs, want 1", len(vals))
	}
	wallet, ok := vals[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("erc8004: getAgentWallet output is %T, want common.Address", vals[0])
	}
	return wallet, nil
}

// OwnerOf reads the ERC-721 owner of an agent id (token id).
func (c *IdentityRegistryClient) OwnerOf(ctx context.Context, tokenID *big.Int) (common.Address, error) {
	out, err := c.callView(ctx, "ownerOf", tokenID)
	if err != nil {
		return common.Address{}, err
	}
	vals, err := c.outputs("ownerOf", out)
	if err != nil {
		return common.Address{}, err
	}
	if len(vals) != 1 {
		return common.Address{}, fmt.Errorf("erc8004: ownerOf returned %d outputs, want 1", len(vals))
	}
	owner, ok := vals[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("erc8004: ownerOf output is %T, want common.Address", vals[0])
	}
	return owner, nil
}

// callView packs the function inputs and performs a read-only eth_call
// against the bound contract address.
func (c *IdentityRegistryClient) callView(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	a, err := IdentityRegistryABI()
	if err != nil {
		return nil, fmt.Errorf("erc8004: parse ABI: %w", err)
	}
	// a.Pack prepends the 4-byte method selector to the packed arguments.
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8004: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &c.address, Data: data}
	return c.rpc.CallContract(ctx, msg, nil)
}

// outputs unpacks the raw call result with the method's declared output types.
func (c *IdentityRegistryClient) outputs(methodName string, data []byte) ([]interface{}, error) {
	a, err := IdentityRegistryABI()
	if err != nil {
		return nil, fmt.Errorf("erc8004: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8004: ABI has no method %q", methodName)
	}
	vals, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("erc8004: unpack %s outputs: %w", methodName, err)
	}
	return vals, nil
}
