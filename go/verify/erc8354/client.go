// Package erc8354 implements the ERC-8354 Confidential Agent Policy Verdicts
// SDK. This file holds the Layer 1 contract clients.
package erc8354

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ConfidentialPolicyVerdictInterfaceID is the ERC-165 interface id of
// IConfidentialPolicyVerdict (0xd6da8150).
const ConfidentialPolicyVerdictInterfaceID = "0xd6da8150"

// Domain is the IPolicyDomainRegistry.Domain record.
type Domain struct {
	Registrar  common.Address
	Verifier   common.Address
	ProgramKey common.Hash
	MaxRootAge uint64
	Active     bool
}

// ConfidentialPolicyVerdictClient consumes a confidential policy verdict via
// a deployed IConfidentialPolicyVerdict contract. verify is a read-only view
// call; consume burns the verdict's single-use nullifier and gates execution.
type ConfidentialPolicyVerdictClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs consume transactions; nil for read-only clients
}

// NewConfidentialPolicyVerdictClient creates a client bound to a deployed
// IConfidentialPolicyVerdict contract. key signs the consume broadcast; pass
// nil for a read-only client.
func NewConfidentialPolicyVerdictClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *ConfidentialPolicyVerdictClient {
	return &ConfidentialPolicyVerdictClient{rpc: rpc, address: addr, key: key}
}

// Verify checks a verdict without state change. Returns false on a well-formed
// but invalid verdict, and never reverts on a malformed proof. The supplied
// context governs the read call.
func (c *ConfidentialPolicyVerdictClient) Verify(ctx context.Context, v Verdict, proof []byte) (bool, error) {
	out, err := callView(ctx, c.rpc, c.address, ConfidentialPolicyVerdictABI, "verify", v, proof)
	if err != nil {
		return false, err
	}
	return unpackBool(ConfidentialPolicyVerdictABI, "verify", out)
}

// VerdictDigest returns the EIP-712 digest an executor signs to authorize a
// relayer.
func (c *ConfidentialPolicyVerdictClient) VerdictDigest(ctx context.Context, v Verdict) (common.Hash, error) {
	out, err := callView(ctx, c.rpc, c.address, ConfidentialPolicyVerdictABI, "verdictDigest", v)
	if err != nil {
		return common.Hash{}, err
	}
	return unpackHash(ConfidentialPolicyVerdictABI, "verdictDigest", out)
}

// IsConsumed reports whether the verdict's nullifier has been burned for the
// domain.
func (c *ConfidentialPolicyVerdictClient) IsConsumed(ctx context.Context, domainId, nullifier common.Hash) (bool, error) {
	out, err := callView(ctx, c.rpc, c.address, ConfidentialPolicyVerdictABI, "isConsumed", domainId, nullifier)
	if err != nil {
		return false, err
	}
	return unpackBool(ConfidentialPolicyVerdictABI, "isConsumed", out)
}

// SupportsInterface reports whether the contract advertises
// IConfidentialPolicyVerdict.
func (c *ConfidentialPolicyVerdictClient) SupportsInterface(ctx context.Context) (bool, error) {
	var id [4]byte
	copy(id[:], common.FromHex(ConfidentialPolicyVerdictInterfaceID))
	out, err := callView(ctx, c.rpc, c.address, ConfidentialPolicyVerdictABI, "supportsInterface", id)
	if err != nil {
		return false, err
	}
	return unpackBool(ConfidentialPolicyVerdictABI, "supportsInterface", out)
}

// Consume verifies and burns a verdict directly. The caller must be the
// executor. Returns the mined receipt. The supplied context governs the
// broadcast and the mining wait.
func (c *ConfidentialPolicyVerdictClient) Consume(ctx context.Context, v Verdict, proof []byte) (*types.Receipt, error) {
	return c.transactAndWait(ctx, "consume", v, proof)
}

// ConsumeRelayed verifies and burns a verdict via a relayer. executorAuth is a
// valid EIP-712 signature by the executor over verdictDigest(v). The three
// argument consume overload is stored by go-ethereum as consume0 because the
// ABI carries two consume methods.
func (c *ConfidentialPolicyVerdictClient) ConsumeRelayed(ctx context.Context, v Verdict, proof, executorAuth []byte) (*types.Receipt, error) {
	return c.transactAndWait(ctx, "consume0", v, proof, executorAuth)
}

func (c *ConfidentialPolicyVerdictClient) transactAndWait(ctx context.Context, method string, args ...interface{}) (*types.Receipt, error) {
	if c.key == nil {
		return nil, fmt.Errorf("erc8354: %s requires a signer key", method)
	}
	a, err := ConfidentialPolicyVerdictABI()
	if err != nil {
		return nil, fmt.Errorf("erc8354: parse ABI: %w", err)
	}
	chainID, err := c.rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erc8354: fetch chain id: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.key, chainID)
	if err != nil {
		return nil, fmt.Errorf("erc8354: create transactor: %w", err)
	}
	// The caller-supplied context governs the whole operation, including the
	// nonce, fee and gas RPC calls performed by the transactor during
	// broadcast, not only the mining wait.
	auth.Context = ctx
	bound := bind.NewBoundContract(c.address, a, c.rpc, c.rpc, c.rpc)
	tx, err := bound.Transact(auth, method, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8354: %s: %w", method, err)
	}
	receipt, err := bind.WaitMined(ctx, c.rpc, tx)
	if err != nil {
		// The transaction has already been broadcast; retain its hash so the
		// caller can still track a transaction that may later mine.
		return nil, fmt.Errorf("erc8354: wait for %s to mine (tx %s): %w", method, tx.Hash().Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("erc8354: %s reverted (tx %s)", method, tx.Hash().Hex())
	}
	return receipt, nil
}

// PolicyDomainRegistryClient reads policy domains and roots via a deployed
// IPolicyDomainRegistry contract. All calls are read-only.
type PolicyDomainRegistryClient struct {
	rpc     *ethclient.Client
	address common.Address
}

// NewPolicyDomainRegistryClient creates a client bound to a deployed
// IPolicyDomainRegistry contract.
func NewPolicyDomainRegistryClient(rpc *ethclient.Client, addr common.Address) *PolicyDomainRegistryClient {
	return &PolicyDomainRegistryClient{rpc: rpc, address: addr}
}

// Domain returns the Domain record for a domain id.
func (c *PolicyDomainRegistryClient) Domain(ctx context.Context, domainId common.Hash) (Domain, error) {
	out, err := callView(ctx, c.rpc, c.address, PolicyDomainRegistryABI, "domain", domainId)
	if err != nil {
		return Domain{}, err
	}
	a, err := PolicyDomainRegistryABI()
	if err != nil {
		return Domain{}, err
	}
	method, ok := a.Methods["domain"]
	if !ok {
		return Domain{}, fmt.Errorf("erc8354: ABI has no method domain")
	}
	vals, err := method.Outputs.Unpack(out)
	if err != nil {
		return Domain{}, fmt.Errorf("erc8354: unpack domain outputs: %w", err)
	}
	var decoded struct {
		Domain Domain
	}
	if err := method.Outputs.Copy(&decoded, vals); err != nil {
		return Domain{}, fmt.Errorf("erc8354: copy domain outputs: %w", err)
	}
	return decoded.Domain, nil
}

// CurrentRoot returns the current root, version, and update timestamp for a
// domain id.
func (c *PolicyDomainRegistryClient) CurrentRoot(ctx context.Context, domainId common.Hash) (common.Hash, uint64, uint64, error) {
	out, err := callView(ctx, c.rpc, c.address, PolicyDomainRegistryABI, "currentRoot", domainId)
	if err != nil {
		return common.Hash{}, 0, 0, err
	}
	vals, err := unpackOutputs(PolicyDomainRegistryABI, "currentRoot", out)
	if err != nil {
		return common.Hash{}, 0, 0, err
	}
	if len(vals) != 3 {
		return common.Hash{}, 0, 0, fmt.Errorf("erc8354: currentRoot returned %d outputs, want 3", len(vals))
	}
	root, ok := vals[0].([32]byte)
	if !ok {
		return common.Hash{}, 0, 0, fmt.Errorf("erc8354: currentRoot root is %T, want [32]byte", vals[0])
	}
	version, ok := vals[1].(uint64)
	if !ok {
		return common.Hash{}, 0, 0, fmt.Errorf("erc8354: currentRoot version is %T, want uint64", vals[1])
	}
	updatedAt, ok := vals[2].(uint64)
	if !ok {
		return common.Hash{}, 0, 0, fmt.Errorf("erc8354: currentRoot updatedAt is %T, want uint64", vals[2])
	}
	return common.BytesToHash(root[:]), version, updatedAt, nil
}

// IsRootAcceptable reports whether a root is current or superseded within the
// grace window.
func (c *PolicyDomainRegistryClient) IsRootAcceptable(ctx context.Context, domainId, root common.Hash) (bool, error) {
	out, err := callView(ctx, c.rpc, c.address, PolicyDomainRegistryABI, "isRootAcceptable", domainId, root)
	if err != nil {
		return false, err
	}
	return unpackBool(PolicyDomainRegistryABI, "isRootAcceptable", out)
}

func callView(ctx context.Context, rpc *ethclient.Client, address common.Address, abiGetter func() (abi.ABI, error), methodName string, args ...interface{}) ([]byte, error) {
	a, err := abiGetter()
	if err != nil {
		return nil, fmt.Errorf("erc8354: parse ABI: %w", err)
	}
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8354: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &address, Data: data}
	return rpc.CallContract(ctx, msg, nil)
}

func unpackOutputs(abiGetter func() (abi.ABI, error), methodName string, data []byte) ([]interface{}, error) {
	a, err := abiGetter()
	if err != nil {
		return nil, fmt.Errorf("erc8354: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8354: ABI has no method %q", methodName)
	}
	return method.Outputs.Unpack(data)
}

func unpackBool(abiGetter func() (abi.ABI, error), methodName string, data []byte) (bool, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return false, err
	}
	if len(vals) != 1 {
		return false, fmt.Errorf("erc8354: %s returned %d outputs, want 1", methodName, len(vals))
	}
	v, ok := vals[0].(bool)
	if !ok {
		return false, fmt.Errorf("erc8354: %s output is %T, want bool", methodName, vals[0])
	}
	return v, nil
}

func unpackHash(abiGetter func() (abi.ABI, error), methodName string, data []byte) (common.Hash, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return common.Hash{}, err
	}
	if len(vals) != 1 {
		return common.Hash{}, fmt.Errorf("erc8354: %s returned %d outputs, want 1", methodName, len(vals))
	}
	b, ok := vals[0].([32]byte)
	if !ok {
		return common.Hash{}, fmt.Errorf("erc8354: %s output is %T, want [32]byte", methodName, vals[0])
	}
	return common.BytesToHash(b[:]), nil
}