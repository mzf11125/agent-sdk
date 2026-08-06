package erc8299

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WyriweAttestationClient checks WYRIWE (L3 input provenance) attestations
// against a deployed IWyriweAttestation contract. All calls are read-only
// view functions — no gas or broadcast needed. Anyone can re-derive the
// answer without spending gas or holding a funded key — that's the whole
// point of exposing verification this way.
type WyriweAttestationClient struct {
	rpc     *ethclient.Client
	address common.Address
}

// NewWyriweAttestationClient creates a client bound to a deployed ERC-8299
// IWyriweAttestation contract.
func NewWyriweAttestationClient(rpc *ethclient.Client, addr common.Address) *WyriweAttestationClient {
	return &WyriweAttestationClient{rpc: rpc, address: addr}
}

// Verify checks an attestation's signature against the contract's known
// attestor via the IWyriweAttestation.verify view. For the testkit mock,
// the "signature" is the raw 32-byte keccak256 of abi.encode(attestation);
// a conforming implementation checks the EIP-712 digest. Either way the
// boolean is computed on-chain and never trusted off-chain.
func (c *WyriweAttestationClient) Verify(ctx context.Context, attestation WyriweAttestation, signature []byte) (bool, error) {
	out, err := callView(ctx, c.rpc, c.address, WyriweAttestationABI, "verify", attestation, signature)
	if err != nil {
		return false, err
	}
	return unpackBool(WyriweAttestationABI, "verify", out)
}

// ProofSystem reads the IWyriweAttestation.proofSystem view — always
// "attestation/wyriwe" for conforming implementations.
func (c *WyriweAttestationClient) ProofSystem(ctx context.Context) (string, error) {
	out, err := callView(ctx, c.rpc, c.address, WyriweAttestationABI, "proofSystem")
	if err != nil {
		return "", err
	}
	return unpackString(WyriweAttestationABI, "proofSystem", out)
}

// JudgmentExecutionClient checks judgment execution (L4 chain-of-custody)
// attestations against a deployed IJudgmentExecutionAttestation contract.
// All calls are read-only view functions — no gas or broadcast needed.
type JudgmentExecutionClient struct {
	rpc     *ethclient.Client
	address common.Address
}

// NewJudgmentExecutionClient creates a client bound to a deployed ERC-8299
// IJudgmentExecutionAttestation contract.
func NewJudgmentExecutionClient(rpc *ethclient.Client, addr common.Address) *JudgmentExecutionClient {
	return &JudgmentExecutionClient{rpc: rpc, address: addr}
}

// Verify checks a judgment execution attestation's signature against the
// contract's known executing-agent attestor via the
// IJudgmentExecutionAttestation.verify view. Only one signature is
// required: the executing agent's attestor signs at reveal time.
func (c *JudgmentExecutionClient) Verify(ctx context.Context, attestation JudgmentExecutionAttestation, signature []byte) (bool, error) {
	out, err := callView(ctx, c.rpc, c.address, JudgmentExecutionABI, "verify", attestation, signature)
	if err != nil {
		return false, err
	}
	return unpackBool(JudgmentExecutionABI, "verify", out)
}

// ProofSystem reads the IJudgmentExecutionAttestation.proofSystem view —
// always "attestation/judgment" for conforming implementations.
func (c *JudgmentExecutionClient) ProofSystem(ctx context.Context) (string, error) {
	out, err := callView(ctx, c.rpc, c.address, JudgmentExecutionABI, "proofSystem")
	if err != nil {
		return "", err
	}
	return unpackString(JudgmentExecutionABI, "proofSystem", out)
}

// callView packs the function inputs and performs a read-only eth_call
// against the bound contract address.
//
// GOTCHA: the inputs must be packed via abi.ABI.Pack(name, args...), which
// prepends the 4-byte method selector — packing the arguments standalone
// with abi.Arguments.Pack produces calldata with no selector, which the
// node would reject. The tuple struct argument is encoded by go-ethereum
// from the exported struct fields (matched to the ABI component names).
func callView(ctx context.Context, rpc *ethclient.Client, address common.Address, abiGetter func() (abi.ABI, error), methodName string, args ...interface{}) ([]byte, error) {
	a, err := abiGetter()
	if err != nil {
		return nil, fmt.Errorf("erc8299: parse ABI: %w", err)
	}
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8299: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &address, Data: data}
	return rpc.CallContract(ctx, msg, nil)
}

// unpackBool decodes a single-boolean view result.
func unpackBool(abiGetter func() (abi.ABI, error), methodName string, data []byte) (bool, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return false, err
	}
	if len(vals) != 1 {
		return false, fmt.Errorf("erc8299: %s returned %d outputs, want 1", methodName, len(vals))
	}
	v, ok := vals[0].(bool)
	if !ok {
		return false, fmt.Errorf("erc8299: %s output is %T, want bool", methodName, vals[0])
	}
	return v, nil
}

// unpackString decodes a single-string view result.
func unpackString(abiGetter func() (abi.ABI, error), methodName string, data []byte) (string, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return "", err
	}
	if len(vals) != 1 {
		return "", fmt.Errorf("erc8299: %s returned %d outputs, want 1", methodName, len(vals))
	}
	s, ok := vals[0].(string)
	if !ok {
		return "", fmt.Errorf("erc8299: %s output is %T, want string", methodName, vals[0])
	}
	return s, nil
}

// unpackOutputs unpacks the raw call result with the method's declared
// output types.
func unpackOutputs(abiGetter func() (abi.ABI, error), methodName string, data []byte) ([]interface{}, error) {
	a, err := abiGetter()
	if err != nil {
		return nil, fmt.Errorf("erc8299: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8299: ABI has no method %q", methodName)
	}
	vals, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("erc8299: unpack %s outputs: %w", methodName, err)
	}
	return vals, nil
}
