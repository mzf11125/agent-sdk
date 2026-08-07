package erc8275

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// AgentReputationClient reads reputation state from a deployed ERC-8275
// contract. All calls are read-only view functions — no gas or broadcast
// needed.
type AgentReputationClient struct {
	rpc     *ethclient.Client
	address common.Address
}

// NewAgentReputationClient creates a client bound to a deployed ERC-8275
// contract.
func NewAgentReputationClient(rpc *ethclient.Client, addr common.Address) *AgentReputationClient {
	return &AgentReputationClient{rpc: rpc, address: addr}
}

// GetReputation reads the current reputation snapshot for an agent.
//
// Returns ReputationData with completed/disputed orders, volume, last active
// timestamp, and score.
func (c *AgentReputationClient) GetReputation(ctx context.Context, agentID common.Hash) (ReputationData, error) {
	out, err := c.callView(ctx, "getReputation", agentID)
	if err != nil {
		return ReputationData{}, err
	}
	vals, err := c.outputs("getReputation", out)
	if err != nil {
		return ReputationData{}, err
	}
	if len(vals) != 5 {
		return ReputationData{}, fmt.Errorf("erc8275: getReputation returned %d outputs, want 5", len(vals))
	}
	completed, err := asUint64(vals[0])
	if err != nil {
		return ReputationData{}, err
	}
	disputed, err := asUint64(vals[1])
	if err != nil {
		return ReputationData{}, err
	}
	volume, err := asUint64(vals[2])
	if err != nil {
		return ReputationData{}, err
	}
	lastActive, err := asUint64(vals[3])
	if err != nil {
		return ReputationData{}, err
	}
	score, err := asUint16(vals[4])
	if err != nil {
		return ReputationData{}, err
	}
	return ReputationData{
		CompletedOrders: completed,
		DisputedOrders:  disputed,
		TotalVolume:     volume,
		LastActiveAt:    lastActive,
		Score:           score,
	}, nil
}

// GetDecayWeight reads the recency-decay weight applied to an agent's score.
//
// Returns the decay weight in basis points (10000 = no decay).
func (c *AgentReputationClient) GetDecayWeight(ctx context.Context, agentID common.Hash) (uint16, error) {
	out, err := c.callView(ctx, "getDecayWeight", agentID)
	if err != nil {
		return 0, err
	}
	vals, err := c.outputs("getDecayWeight", out)
	if err != nil {
		return 0, err
	}
	if len(vals) != 1 {
		return 0, fmt.Errorf("erc8275: getDecayWeight returned %d outputs, want 1", len(vals))
	}
	return asUint16(vals[0])
}

// VerifyOutcome verifies a settled order's outcome proof against the public
// record. A read-only simulated call, not a broadcast transaction — anyone
// can freely re-derive the answer without spending gas or holding a key.
//
// Returns true if the outcome is valid against public on-chain data.
func (c *AgentReputationClient) VerifyOutcome(ctx context.Context, orderID common.Hash, proof []byte) (bool, error) {
	out, err := c.callView(ctx, "verifyOutcome", orderID, proof)
	if err != nil {
		return false, err
	}
	vals, err := c.outputs("verifyOutcome", out)
	if err != nil {
		return false, err
	}
	if len(vals) != 1 {
		return false, fmt.Errorf("erc8275: verifyOutcome returned %d outputs, want 1", len(vals))
	}
	valid, ok := vals[0].(bool)
	if !ok {
		return false, fmt.Errorf("erc8275: verifyOutcome output is %T, want bool", vals[0])
	}
	return valid, nil
}

// callView packs the function inputs and performs a read-only eth_call
// against the bound contract address.
func (c *AgentReputationClient) callView(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	a, err := AgentReputationABI()
	if err != nil {
		return nil, fmt.Errorf("erc8275: parse ABI: %w", err)
	}
	// a.Pack prepends the 4-byte method selector to the packed arguments.
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8275: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &c.address, Data: data}
	return c.rpc.CallContract(ctx, msg, nil)
}

// outputs unpacks the raw call result with the method's declared output types.
func (c *AgentReputationClient) outputs(methodName string, data []byte) ([]interface{}, error) {
	a, err := AgentReputationABI()
	if err != nil {
		return nil, fmt.Errorf("erc8275: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8275: ABI has no method %q", methodName)
	}
	vals, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("erc8275: unpack %s outputs: %w", methodName, err)
	}
	return vals, nil
}

func asUint64(v interface{}) (uint64, error) {
	if u, ok := v.(uint64); ok {
		return u, nil
	}
	return 0, fmt.Errorf("erc8275: expected uint64 output, got %T", v)
}

func asUint16(v interface{}) (uint16, error) {
	if u, ok := v.(uint16); ok {
		return u, nil
	}
	return 0, fmt.Errorf("erc8275: expected uint16 output, got %T", v)
}
