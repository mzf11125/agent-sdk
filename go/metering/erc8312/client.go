package erc8312

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ErrNoSigner is returned by the write methods (RegisterEnvelope,
// AdvanceCursor, SetStatus, Contest, Resolve) when the client was
// constructed without a private key.
var ErrNoSigner = errors.New("erc8312: write methods require a signer key (non-nil key in the client constructor)")

// BoundedAgentActionClient reads and drives envelope lifecycle state on a
// deployed ERC-8312 IBoundedAgentAction contract. GetEnvelope/GetCursor/
// GetStatus/IsActive are read-only view calls (no gas, no key);
// RegisterEnvelope/AdvanceCursor/SetStatus broadcast transactions and
// therefore need a signer key — pass nil to NewBoundedAgentActionClient for
// a read-only client.
type BoundedAgentActionClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs registerEnvelope/advanceCursor/setStatus; nil for read-only clients
}

// NewBoundedAgentActionClient creates a client bound to a deployed ERC-8312
// IBoundedAgentAction contract. key signs the write transactions — pass nil
// for a read-only client (GetEnvelope, GetCursor, GetStatus, IsActive only).
func NewBoundedAgentActionClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *BoundedAgentActionClient {
	return &BoundedAgentActionClient{rpc: rpc, address: addr, key: key}
}

// RegisterEnvelope registers a new envelope by broadcasting
// registerEnvelope(principal, capabilityRoot, expiresAt, initData), waits
// for the transaction to be mined, and returns the contract-generated id
// parsed from the EnvelopeRegistered event. Returns ErrNoSigner if the
// client has no private key.
func (c *BoundedAgentActionClient) RegisterEnvelope(ctx context.Context, principal common.Address, capabilityRoot common.Hash, expiresAt uint64, initData []byte) (common.Hash, error) {
	if c.key == nil {
		return common.Hash{}, ErrNoSigner
	}
	receipt, err := c.transactAndWait(ctx, "registerEnvelope", principal, capabilityRoot, expiresAt, initData)
	if err != nil {
		return common.Hash{}, err
	}
	return parseEnvelopeRegistered(receipt)
}

// AdvanceCursor advances the envelope's cursor with a substrate-validated
// witness by broadcasting advanceCursor(id, witness), waits for the
// transaction to be mined, and returns the previous and new cursor
// commitments parsed from the EnvelopeAdvanced event. Returns ErrNoSigner if
// the client has no private key.
func (c *BoundedAgentActionClient) AdvanceCursor(ctx context.Context, id common.Hash, witness []byte) (AdvanceResult, error) {
	if c.key == nil {
		return AdvanceResult{}, ErrNoSigner
	}
	receipt, err := c.transactAndWait(ctx, "advanceCursor", id, witness)
	if err != nil {
		return AdvanceResult{}, err
	}
	return parseEnvelopeAdvanced(receipt)
}

// SetStatus transitions the envelope's lifecycle status by broadcasting
// setStatus(id, newStatus). Returns ErrNoSigner if the client has no private
// key; the caller may wait on the returned transaction with bind.WaitMined.
func (c *BoundedAgentActionClient) SetStatus(ctx context.Context, id common.Hash, newStatus Status) (*types.Transaction, error) {
	if c.key == nil {
		return nil, ErrNoSigner
	}
	return c.transact(ctx, "setStatus", id, uint8(newStatus))
}

// GetEnvelope reads the full stored Envelope for id via the getEnvelope
// view. The stored struct is the recompute-to-verify input: the metering
// invariants (see CheckStatefulBound, CheckCursorHeadroom) are pure
// functions of its fields. Errors if the contract reverts (unknown id).
func (c *BoundedAgentActionClient) GetEnvelope(ctx context.Context, id common.Hash) (Envelope, error) {
	out, err := c.callView(ctx, "getEnvelope", id)
	if err != nil {
		return Envelope{}, err
	}
	vals, err := c.outputs("getEnvelope", out)
	if err != nil {
		return Envelope{}, err
	}
	if len(vals) != 1 {
		return Envelope{}, fmt.Errorf("erc8312: getEnvelope returned %d outputs, want 1", len(vals))
	}
	// Outputs.Copy maps the tuple output onto the anonymous struct by
	// component name (id, principal, capabilityRoot, cursorRoot, createdAt,
	// expiresAt, status). The Status component must be a plain uint8 here:
	// Copy requires exact type matches and cannot unmarshal into the named
	// Status type — the conversion happens below.
	var decoded struct {
		Envelope struct {
			Id             [32]byte
			Principal      common.Address
			CapabilityRoot [32]byte
			CursorRoot     [32]byte
			CreatedAt      uint64
			ExpiresAt      uint64
			Status         uint8
		}
	}
	if err := c.copyOutputs("getEnvelope", &decoded, vals); err != nil {
		return Envelope{}, err
	}
	raw := decoded.Envelope
	return Envelope{
		Id:             common.BytesToHash(raw.Id[:]),
		Principal:      raw.Principal,
		CapabilityRoot: common.BytesToHash(raw.CapabilityRoot[:]),
		CursorRoot:     common.BytesToHash(raw.CursorRoot[:]),
		CreatedAt:      raw.CreatedAt,
		ExpiresAt:      raw.ExpiresAt,
		Status:         Status(raw.Status),
	}, nil
}

// GetCursor reads the current cursor commitment via the getCursor view.
// MUST equal getEnvelope(id).cursorRoot.
func (c *BoundedAgentActionClient) GetCursor(ctx context.Context, id common.Hash) (common.Hash, error) {
	out, err := c.callView(ctx, "getCursor", id)
	if err != nil {
		return common.Hash{}, err
	}
	return c.singleHash("getCursor", out)
}

// GetStatus reads the effective envelope status via the getStatus view
// (reports Expired once expiresAt is reached, even while the stored status
// remains Active).
func (c *BoundedAgentActionClient) GetStatus(ctx context.Context, id common.Hash) (Status, error) {
	out, err := c.callView(ctx, "getStatus", id)
	if err != nil {
		return 0, err
	}
	vals, err := c.outputs("getStatus", out)
	if err != nil {
		return 0, err
	}
	if len(vals) != 1 {
		return 0, fmt.Errorf("erc8312: getStatus returned %d outputs, want 1", len(vals))
	}
	v, ok := vals[0].(uint8)
	if !ok {
		return 0, fmt.Errorf("erc8312: getStatus output is %T, want uint8", vals[0])
	}
	return Status(v), nil
}

// IsActive reports whether the effective status is Active via the isActive
// view.
func (c *BoundedAgentActionClient) IsActive(ctx context.Context, id common.Hash) (bool, error) {
	out, err := c.callView(ctx, "isActive", id)
	if err != nil {
		return false, err
	}
	vals, err := c.outputs("isActive", out)
	if err != nil {
		return false, err
	}
	if len(vals) != 1 {
		return false, fmt.Errorf("erc8312: isActive returned %d outputs, want 1", len(vals))
	}
	v, ok := vals[0].(bool)
	if !ok {
		return false, fmt.Errorf("erc8312: isActive output is %T, want bool", vals[0])
	}
	return v, nil
}

// BudgetSubstrateClient reads the ERC-8312 budget profile views on a
// deployed IBudgetSubstrate contract. All calls are read-only view
// functions — no gas, no signer key needed. The typed accessor and the
// cursor commitment cannot diverge: cursorRoot = keccak256(abi.encode(spent))
// and remaining = cap - spent are recomputable off-chain (see
// ComputeRemainingHeadroom, VerifyRemaining).
type BudgetSubstrateClient struct {
	rpc     *ethclient.Client
	address common.Address
}

// NewBudgetSubstrateClient creates a client bound to a deployed ERC-8312
// IBudgetSubstrate contract.
func NewBudgetSubstrateClient(rpc *ethclient.Client, addr common.Address) *BudgetSubstrateClient {
	return &BudgetSubstrateClient{rpc: rpc, address: addr}
}

// Bound reads the configured bound (cap, asset) for an envelope via the
// bound view. Errors if the contract reverts (unknown id).
func (c *BudgetSubstrateClient) Bound(ctx context.Context, id common.Hash) (Bound, error) {
	out, err := c.callView(ctx, "bound", id)
	if err != nil {
		return Bound{}, err
	}
	vals, err := c.outputs("bound", out)
	if err != nil {
		return Bound{}, err
	}
	if len(vals) != 2 {
		return Bound{}, fmt.Errorf("erc8312: bound returned %d outputs, want 2", len(vals))
	}
	cap, err := asUint64(vals[0])
	if err != nil {
		return Bound{}, err
	}
	asset, err := asAddress(vals[1])
	if err != nil {
		return Bound{}, err
	}
	return Bound{Cap: cap, Asset: asset}, nil
}

// Spent reads the cumulative value consumed under the envelope via the
// spent view. Errors if the contract reverts (unknown id).
func (c *BudgetSubstrateClient) Spent(ctx context.Context, id common.Hash) (uint64, error) {
	out, err := c.callView(ctx, "spent", id)
	if err != nil {
		return 0, err
	}
	return c.singleUint64("spent", out)
}

// Remaining reads the remaining headroom (cap - spent) via the remaining
// view — or 0 if the envelope is not active. The on-chain value is
// independently recomputable off-chain with ComputeRemainingHeadroom and
// cross-checkable with VerifyRemaining.
func (c *BudgetSubstrateClient) Remaining(ctx context.Context, id common.Hash) (uint64, error) {
	out, err := c.callView(ctx, "remaining", id)
	if err != nil {
		return 0, err
	}
	return c.singleUint64("remaining", out)
}

// ContestableEnvelopeClient drives the ERC-8312 contestation lifecycle on a
// deployed IContestableEnvelope contract. Both methods broadcast
// transactions and need a signer key — pass nil to
// NewContestableEnvelopeClient for a read-only client.
type ContestableEnvelopeClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs contest/resolve
}

// NewContestableEnvelopeClient creates a client bound to a deployed ERC-8312
// IContestableEnvelope contract. key signs the write transactions — pass nil
// for a read-only client.
func NewContestableEnvelopeClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *ContestableEnvelopeClient {
	return &ContestableEnvelopeClient{rpc: rpc, address: addr, key: key}
}

// Contest contests an active envelope (Active -> Contested) by broadcasting
// contest(id, evidence), waits for the transaction to be mined, and returns
// the contest info parsed from the EnvelopeContested event. Returns
// ErrNoSigner if the client has no private key.
func (c *ContestableEnvelopeClient) Contest(ctx context.Context, id common.Hash, evidence []byte) (ContestInfo, error) {
	if c.key == nil {
		return ContestInfo{}, ErrNoSigner
	}
	receipt, err := c.transactAndWait(ctx, "contest", id, evidence)
	if err != nil {
		return ContestInfo{}, err
	}
	return parseEnvelopeContested(receipt)
}

// Resolve resolves a contested envelope (Contested -> Active or Revoked) by
// broadcasting resolve(id, outcome, resolution), waits for the transaction
// to be mined, and returns the resolution info parsed from the
// EnvelopeResolved event. Returns ErrNoSigner if the client has no private
// key.
func (c *ContestableEnvelopeClient) Resolve(ctx context.Context, id common.Hash, outcome Status, resolution []byte) (ResolveInfo, error) {
	if c.key == nil {
		return ResolveInfo{}, ErrNoSigner
	}
	receipt, err := c.transactAndWait(ctx, "resolve", id, uint8(outcome), resolution)
	if err != nil {
		return ResolveInfo{}, err
	}
	return parseEnvelopeResolved(receipt)
}

// transact packs the method inputs via the ABI (a.Pack prepends the 4-byte
// method selector), signs with the client's key and broadcasts. Gas limit,
// base fee and nonce are resolved against the live node; the chain id is
// fetched from the RPC at call time.
func (c *BoundedAgentActionClient) transact(ctx context.Context, methodName string, args ...interface{}) (*types.Transaction, error) {
	a, err := BoundedAgentActionABI()
	if err != nil {
		return nil, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	return transactWith(c.rpc, c.key, c.address, a, methodName, args...)
}

// transactAndWait broadcasts a transaction and waits for it to be mined,
// returning an error if the transaction reverted.
func (c *BoundedAgentActionClient) transactAndWait(ctx context.Context, methodName string, args ...interface{}) (*types.Receipt, error) {
	tx, err := c.transact(ctx, methodName, args...)
	if err != nil {
		return nil, err
	}
	return waitMinedChecked(ctx, c.rpc, tx, methodName)
}

// transact packs the method inputs via the ABI, signs with the client's key
// and broadcasts. The method must exist in the contestable ABI.
func (c *ContestableEnvelopeClient) transact(ctx context.Context, methodName string, args ...interface{}) (*types.Transaction, error) {
	a, err := ContestableEnvelopeABI()
	if err != nil {
		return nil, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	return transactWith(c.rpc, c.key, c.address, a, methodName, args...)
}

// transactAndWait broadcasts a transaction and waits for it to be mined,
// returning an error if the transaction reverted.
func (c *ContestableEnvelopeClient) transactAndWait(ctx context.Context, methodName string, args ...interface{}) (*types.Receipt, error) {
	tx, err := c.transact(ctx, methodName, args...)
	if err != nil {
		return nil, err
	}
	return waitMinedChecked(ctx, c.rpc, tx, methodName)
}

// callView packs the function inputs and performs a read-only eth_call
// against the bound contract address.
//
// GOTCHA: the inputs must be packed via abi.ABI.Pack(name, args...), which
// prepends the 4-byte method selector — packing the arguments standalone
// with abi.Arguments.Pack produces calldata with no selector, which the
// node would reject.
func (c *BoundedAgentActionClient) callView(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	a, err := BoundedAgentActionABI()
	if err != nil {
		return nil, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	return callViewWith(ctx, c.rpc, c.address, a, methodName, args...)
}

// outputs unpacks the raw call result with the method's declared output
// types.
func (c *BoundedAgentActionClient) outputs(methodName string, data []byte) ([]interface{}, error) {
	a, err := BoundedAgentActionABI()
	if err != nil {
		return nil, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	return unpackOutputs(a, methodName, data)
}

// copyOutputs copies the decoded outputs into the destination struct via
// abi.Arguments.Copy, mapping every output onto a struct field by name
// (case-insensitive): tuple outputs land in matching struct fields (see
// GetEnvelope).
func (c *BoundedAgentActionClient) copyOutputs(methodName string, dest interface{}, vals []interface{}) error {
	a, err := BoundedAgentActionABI()
	if err != nil {
		return fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return fmt.Errorf("erc8312: ABI has no method %q", methodName)
	}
	if err := method.Outputs.Copy(dest, vals); err != nil {
		return fmt.Errorf("erc8312: copy %s outputs: %w", methodName, err)
	}
	return nil
}

// callView packs the function inputs and performs a read-only eth_call
// against the bound contract address.
func (c *BudgetSubstrateClient) callView(ctx context.Context, methodName string, args ...interface{}) ([]byte, error) {
	a, err := BudgetSubstrateABI()
	if err != nil {
		return nil, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	return callViewWith(ctx, c.rpc, c.address, a, methodName, args...)
}

// outputs unpacks the raw call result with the method's declared output
// types.
func (c *BudgetSubstrateClient) outputs(methodName string, data []byte) ([]interface{}, error) {
	a, err := BudgetSubstrateABI()
	if err != nil {
		return nil, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	return unpackOutputs(a, methodName, data)
}

// singleUint64 decodes a single-uint256 view result (with uint64 range
// check).
func (c *BudgetSubstrateClient) singleUint64(methodName string, data []byte) (uint64, error) {
	vals, err := c.outputs(methodName, data)
	if err != nil {
		return 0, err
	}
	if len(vals) != 1 {
		return 0, fmt.Errorf("erc8312: %s returned %d outputs, want 1", methodName, len(vals))
	}
	return asUint64(vals[0])
}

// singleHash decodes a single-bytes32 view result.
func (c *BoundedAgentActionClient) singleHash(methodName string, data []byte) (common.Hash, error) {
	vals, err := c.outputs(methodName, data)
	if err != nil {
		return common.Hash{}, err
	}
	if len(vals) != 1 {
		return common.Hash{}, fmt.Errorf("erc8312: %s returned %d outputs, want 1", methodName, len(vals))
	}
	return asHash(vals[0])
}

// transactWith signs and broadcasts args packed for methodName on the given
// ABI against address.
func transactWith(rpc *ethclient.Client, key *ecdsa.PrivateKey, address common.Address, a abi.ABI, methodName string, args ...interface{}) (*types.Transaction, error) {
	chainID, err := rpc.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("erc8312: fetch chain id: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, fmt.Errorf("erc8312: create transactor: %w", err)
	}
	bound := bind.NewBoundContract(address, a, rpc, rpc, rpc)
	tx, err := bound.Transact(auth, methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8312: %s: %w", methodName, err)
	}
	return tx, nil
}

// waitMinedChecked waits for the transaction to be mined and returns an
// error if it reverted.
func waitMinedChecked(ctx context.Context, rpc *ethclient.Client, tx *types.Transaction, methodName string) (*types.Receipt, error) {
	receipt, err := bind.WaitMined(ctx, rpc, tx)
	if err != nil {
		return nil, fmt.Errorf("erc8312: wait for %s to mine: %w", methodName, err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("erc8312: %s reverted (tx %s)", methodName, tx.Hash().Hex())
	}
	return receipt, nil
}

// callViewWith packs the function inputs and performs a read-only eth_call
// against the given contract address.
func callViewWith(ctx context.Context, rpc *ethclient.Client, address common.Address, a abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8312: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &address, Data: data}
	return rpc.CallContract(ctx, msg, nil)
}

// unpackOutputs unpacks the raw call result with the method's declared
// output types.
func unpackOutputs(a abi.ABI, methodName string, data []byte) ([]interface{}, error) {
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8312: ABI has no method %q", methodName)
	}
	vals, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("erc8312: unpack %s outputs: %w", methodName, err)
	}
	return vals, nil
}

// parseEnvelopeRegistered extracts the id from an EnvelopeRegistered event
// log. All three inputs are indexed, so they live in the log's topics
// (topic 0 = event signature; topics 1-3 = id, principal, capabilityRoot).
func parseEnvelopeRegistered(receipt *types.Receipt) (common.Hash, error) {
	a, err := BoundedAgentActionABI()
	if err != nil {
		return common.Hash{}, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	evt, ok := a.Events["EnvelopeRegistered"]
	if !ok {
		return common.Hash{}, fmt.Errorf("erc8312: ABI has no event EnvelopeRegistered")
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] != evt.ID {
			continue
		}
		if len(log.Topics) != 4 {
			return common.Hash{}, fmt.Errorf("erc8312: EnvelopeRegistered log has %d topics, want 4", len(log.Topics))
		}
		return common.BytesToHash(log.Topics[1].Bytes()), nil
	}
	return common.Hash{}, fmt.Errorf("erc8312: EnvelopeRegistered event not found in receipt")
}

// parseEnvelopeAdvanced extracts prevCursor and newCursor from an
// EnvelopeAdvanced event log. The id is indexed (topic 1); both cursors are
// non-indexed bytes32 in the log data.
func parseEnvelopeAdvanced(receipt *types.Receipt) (AdvanceResult, error) {
	a, err := BoundedAgentActionABI()
	if err != nil {
		return AdvanceResult{}, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	evt, ok := a.Events["EnvelopeAdvanced"]
	if !ok {
		return AdvanceResult{}, fmt.Errorf("erc8312: ABI has no event EnvelopeAdvanced")
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] != evt.ID {
			continue
		}
		if len(log.Topics) != 2 {
			return AdvanceResult{}, fmt.Errorf("erc8312: EnvelopeAdvanced log has %d topics, want 2", len(log.Topics))
		}
		vals, err := evt.Inputs.NonIndexed().Unpack(log.Data)
		if err != nil {
			return AdvanceResult{}, fmt.Errorf("erc8312: unpack EnvelopeAdvanced data: %w", err)
		}
		if len(vals) != 2 {
			return AdvanceResult{}, fmt.Errorf("erc8312: EnvelopeAdvanced data has %d values, want 2", len(vals))
		}
		prev, ok := vals[0].([32]byte)
		if !ok {
			return AdvanceResult{}, fmt.Errorf("erc8312: EnvelopeAdvanced prevCursor is %T, want [32]byte", vals[0])
		}
		next, ok := vals[1].([32]byte)
		if !ok {
			return AdvanceResult{}, fmt.Errorf("erc8312: EnvelopeAdvanced newCursor is %T, want [32]byte", vals[1])
		}
		return AdvanceResult{
			PrevCursor: common.BytesToHash(prev[:]),
			NewCursor:  common.BytesToHash(next[:]),
		}, nil
	}
	return AdvanceResult{}, fmt.Errorf("erc8312: EnvelopeAdvanced event not found in receipt")
}

// parseEnvelopeContested extracts the challenger from an EnvelopeContested
// event log. Both inputs are indexed (topics 1-2: id, challenger).
func parseEnvelopeContested(receipt *types.Receipt) (ContestInfo, error) {
	a, err := ContestableEnvelopeABI()
	if err != nil {
		return ContestInfo{}, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	evt, ok := a.Events["EnvelopeContested"]
	if !ok {
		return ContestInfo{}, fmt.Errorf("erc8312: ABI has no event EnvelopeContested")
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] != evt.ID {
			continue
		}
		if len(log.Topics) != 3 {
			return ContestInfo{}, fmt.Errorf("erc8312: EnvelopeContested log has %d topics, want 3", len(log.Topics))
		}
		return ContestInfo{
			Id:         common.BytesToHash(log.Topics[1].Bytes()),
			Challenger: common.BytesToAddress(log.Topics[2].Bytes()),
		}, nil
	}
	return ContestInfo{}, fmt.Errorf("erc8312: EnvelopeContested event not found in receipt")
}

// parseEnvelopeResolved extracts the outcome from an EnvelopeResolved event
// log. The id is indexed (topic 1); the outcome is a non-indexed uint8 in
// the log data.
func parseEnvelopeResolved(receipt *types.Receipt) (ResolveInfo, error) {
	a, err := ContestableEnvelopeABI()
	if err != nil {
		return ResolveInfo{}, fmt.Errorf("erc8312: parse ABI: %w", err)
	}
	evt, ok := a.Events["EnvelopeResolved"]
	if !ok {
		return ResolveInfo{}, fmt.Errorf("erc8312: ABI has no event EnvelopeResolved")
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] != evt.ID {
			continue
		}
		if len(log.Topics) != 2 {
			return ResolveInfo{}, fmt.Errorf("erc8312: EnvelopeResolved log has %d topics, want 2", len(log.Topics))
		}
		vals, err := evt.Inputs.NonIndexed().Unpack(log.Data)
		if err != nil {
			return ResolveInfo{}, fmt.Errorf("erc8312: unpack EnvelopeResolved data: %w", err)
		}
		if len(vals) != 1 {
			return ResolveInfo{}, fmt.Errorf("erc8312: EnvelopeResolved data has %d values, want 1", len(vals))
		}
		outcome, ok := vals[0].(uint8)
		if !ok {
			return ResolveInfo{}, fmt.Errorf("erc8312: EnvelopeResolved outcome is %T, want uint8", vals[0])
		}
		return ResolveInfo{Id: common.BytesToHash(log.Topics[1].Bytes()), Outcome: Status(outcome)}, nil
	}
	return ResolveInfo{}, fmt.Errorf("erc8312: EnvelopeResolved event not found in receipt")
}

func asUint64(v interface{}) (uint64, error) {
	switch x := v.(type) {
	case uint64:
		return x, nil
	case *big.Int:
		if !x.IsUint64() {
			return 0, fmt.Errorf("erc8312: uint256 output overflows uint64: %s", x.String())
		}
		return x.Uint64(), nil
	}
	return 0, fmt.Errorf("erc8312: expected uint64 output, got %T", v)
}

func asAddress(v interface{}) (common.Address, error) {
	if a, ok := v.(common.Address); ok {
		return a, nil
	}
	return common.Address{}, fmt.Errorf("erc8312: expected address output, got %T", v)
}

func asHash(v interface{}) (common.Hash, error) {
	switch h := v.(type) {
	case common.Hash:
		return h, nil
	case [32]byte:
		return common.BytesToHash(h[:]), nil
	}
	return common.Hash{}, fmt.Errorf("erc8312: expected bytes32 output, got %T", v)
}
