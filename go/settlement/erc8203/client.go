package erc8203

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

// ErrNoSigner is returned by Resolve when the client was constructed
// without a private key.
var ErrNoSigner = errors.New("erc8203: Resolve requires a signer key (NewConsultEscrowClient with non-nil key)")

// ConsultEscrowClient reads and settles escrowed consultation jobs on a
// deployed ERC-8203 IConsultEscrow contract. GetJob is a read-only view
// call (no gas, no key); Resolve broadcasts a release() transaction and
// therefore needs a signer key — pass nil to NewConsultEscrowClient for a
// read-only client.
type ConsultEscrowClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs release() transactions; nil for read-only clients
}

// NewConsultEscrowClient creates a client bound to a deployed ERC-8203
// IConsultEscrow contract. key signs the release() transaction — pass nil
// for a read-only client (GetJob only).
func NewConsultEscrowClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *ConsultEscrowClient {
	return &ConsultEscrowClient{rpc: rpc, address: addr, key: key}
}

// GetJob reads the escrowed job for jobId via the jobs(bytes32) view.
// A never-opened jobId reads back as the all-zero default (zero addresses,
// zero amount/deadline, Status == JobStatusNone).
func (c *ConsultEscrowClient) GetJob(ctx context.Context, jobID common.Hash) (Job, error) {
	out, err := c.callView(ctx, "jobs", jobID)
	if err != nil {
		return Job{}, err
	}
	vals, err := c.outputs("jobs", out)
	if err != nil {
		return Job{}, err
	}
	if len(vals) != 6 {
		return Job{}, fmt.Errorf("erc8203: jobs returned %d outputs, want 6", len(vals))
	}
	consumer, err := asAddress(vals[0])
	if err != nil {
		return Job{}, err
	}
	provider, err := asAddress(vals[1])
	if err != nil {
		return Job{}, err
	}
	attestor, err := asAddress(vals[2])
	if err != nil {
		return Job{}, err
	}
	amount, err := asBigInt(vals[3])
	if err != nil {
		return Job{}, err
	}
	deadline, err := asBigInt(vals[4])
	if err != nil {
		return Job{}, err
	}
	status, err := asStatus(vals[5])
	if err != nil {
		return Job{}, err
	}
	return Job{
		Consumer: consumer,
		Provider: provider,
		Attestor: attestor,
		Amount:   amount,
		Deadline: deadline,
		Status:   status,
	}, nil
}

// Resolve releases the escrow for jobId to the provider by broadcasting
// release(jobId, resultHash, signature).
//
// resultHash must be keccak256(utf8(resultText)) — the contract recomputes
// commitmentHash = keccak256(abi.encode(jobId, resultHash)) on-chain (see
// ComputeVerdictHash) and only releases if the EIP-191 personal_sign
// over that commitment recovers to the job's attestor. The signature is
// the attestor's 65-byte [r || s || v] (v = 27 or 28), not the transaction
// signature.
//
// The returned transaction is signed with the client's key and broadcast;
// the caller can wait for its receipt (bind.WaitMined). Gas limit, base
// fee and nonce are resolved against the live node; the chain id is
// fetched from the RPC at call time. Returns ErrNoSigner if the client has
// no private key.
func (c *ConsultEscrowClient) Resolve(ctx context.Context, jobID common.Hash, resultHash common.Hash, signature []byte) (*types.Transaction, error) {
	if c.key == nil {
		return nil, ErrNoSigner
	}
	a, err := ConsultEscrowABI()
	if err != nil {
		return nil, fmt.Errorf("erc8203: parse ABI: %w", err)
	}
	chainID, err := c.rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erc8203: fetch chain id: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.key, chainID)
	if err != nil {
		return nil, fmt.Errorf("erc8203: create transactor: %w", err)
	}
	// Transact packs the inputs via a.Pack("release", jobId, resultHash,
	// signature) — which prepends the 4-byte method selector — signs,
	// estimates gas (GasLimit 0) and broadcasts through the same rpc client.
	bound := bind.NewBoundContract(c.address, a, c.rpc, c.rpc, c.rpc)
	tx, err := bound.Transact(auth, "release", jobID, resultHash, signature)
	if err != nil {
		return nil, fmt.Errorf("erc8203: release(%s, %s): %w", jobID.Hex(), resultHash.Hex(), err)
	}
	return tx, nil
}

// callView packs the function inputs and performs a read-only eth_call
// against the bound contract address.
//
// GOTCHA: the inputs must be packed via abi.ABI.Pack(name, args...), which
// prepends the 4-byte method selector — packing the arguments standalone
// with abi.Arguments.Pack produces calldata with no selector, which the
// node would reject.
func (c *ConsultEscrowClient) callView(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	a, err := ConsultEscrowABI()
	if err != nil {
		return nil, fmt.Errorf("erc8203: parse ABI: %w", err)
	}
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8203: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &c.address, Data: data}
	return c.rpc.CallContract(ctx, msg, nil)
}

// outputs unpacks the raw call result with the method's declared output
// types.
func (c *ConsultEscrowClient) outputs(methodName string, data []byte) ([]interface{}, error) {
	a, err := ConsultEscrowABI()
	if err != nil {
		return nil, fmt.Errorf("erc8203: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8203: ABI has no method %q", methodName)
	}
	vals, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("erc8203: unpack %s outputs: %w", methodName, err)
	}
	return vals, nil
}

func asAddress(v interface{}) (common.Address, error) {
	if a, ok := v.(common.Address); ok {
		return a, nil
	}
	return common.Address{}, fmt.Errorf("erc8203: expected address output, got %T", v)
}

func asBigInt(v interface{}) (*big.Int, error) {
	if i, ok := v.(*big.Int); ok {
		return i, nil
	}
	return nil, fmt.Errorf("erc8203: expected *big.Int output, got %T", v)
}

func asStatus(v interface{}) (JobStatus, error) {
	if s, ok := v.(uint8); ok {
		return JobStatus(s), nil
	}
	return 0, fmt.Errorf("erc8203: expected uint8 status output, got %T", v)
}
