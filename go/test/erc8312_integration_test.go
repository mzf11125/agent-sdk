// Package test holds integration tests that run against a local anvil node
// deployed via the testkit harness.
//
// Usage:
//
//	testkit/scripts/start-anvil.sh
//	ERC8312_ADDRESSES=$(testkit/scripts/deploy.sh metering/ERC8312 DeployERC8312)
//	ERC8312_ADDRESSES="$ERC8312_ADDRESSES" go test -v ./go/test/ -run TestERC8312
//	testkit/scripts/stop-anvil.sh
//
// deploy.sh prints one contract-creation address per line in broadcast
// order: the boundedAgentAction mock first, then budgetSubstrate, then
// contestableEnvelope. ERC8312_ADDRESSES carries those three addresses
// separated by newlines (shell command substitution preserves internal
// newlines).
package test

import (
	"context"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/metering/erc8312"
)

const (
	erc8312ExpiresAt   = 2000000000 // far future — envelope stays Active
	erc8312BudgetCap   = 10000      // MockBudgetSubstrate DEFAULT_CAP
	erc8312BudgetDelta = 500        // first advance on the budget substrate
)

func TestERC8312BoundedAgentAction(t *testing.T) {
	boundedAddr, _, _ := erc8312Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key := anvilSignerKey(t)
	signer := crypto.PubkeyToAddress(key.PublicKey)
	client := erc8312.NewBoundedAgentActionClient(rpc, boundedAddr, key)

	// ── registerEnvelope(): reserve a mandate under a capability root ──────
	capRoot := crypto.Keccak256Hash([]byte("my-capability"))
	id, err := client.RegisterEnvelope(context.Background(), signer, capRoot, erc8312ExpiresAt, nil)
	if err != nil {
		t.Fatalf("RegisterEnvelope: %v", err)
	}
	t.Logf("registerEnvelope → id=%s", id.Hex())
	if id == (common.Hash{}) {
		t.Error("RegisterEnvelope returned zero id")
	}

	// ── getEnvelope(): the stored struct is the recompute-to-verify input ──
	env, err := client.GetEnvelope(context.Background(), id)
	if err != nil {
		t.Fatalf("GetEnvelope(%s): %v", id.Hex(), err)
	}
	t.Logf("getEnvelope → principal=%s capabilityRoot=%s cursorRoot=%s status=%d",
		env.Principal.Hex(), env.CapabilityRoot.Hex(), env.CursorRoot.Hex(), env.Status)
	if env.Principal != signer {
		t.Errorf("envelope principal = %s, want signer %s", env.Principal.Hex(), signer.Hex())
	}
	if env.CapabilityRoot != capRoot {
		t.Errorf("envelope capabilityRoot = %s, want %s", env.CapabilityRoot.Hex(), capRoot.Hex())
	}
	if env.Status != erc8312.StatusActive {
		t.Errorf("envelope status = %d, want %d (Active)", env.Status, erc8312.StatusActive)
	}
	if env.CreatedAt == 0 {
		t.Error("envelope createdAt = 0, want block.timestamp")
	}

	// ── isActive() ──────────────────────────────────────────────────────────
	active, err := client.IsActive(context.Background(), id)
	if err != nil {
		t.Fatalf("IsActive: %v", err)
	}
	if !active {
		t.Error("isActive = false right after registration, want true")
	}

	// ── advanceCursor(): metering consumes — the witness is public ──────────
	adv, err := client.AdvanceCursor(context.Background(), id, []byte("witness-data"))
	if err != nil {
		t.Fatalf("AdvanceCursor: %v", err)
	}
	t.Logf("advanceCursor → prev=%s new=%s", adv.PrevCursor.Hex(), adv.NewCursor.Hex())
	if adv.NewCursor == (common.Hash{}) {
		t.Error("advanceCursor returned zero newCursor")
	}
	cursor, err := client.GetCursor(context.Background(), id)
	if err != nil {
		t.Fatalf("GetCursor: %v", err)
	}
	if cursor != adv.NewCursor {
		t.Errorf("getCursor = %s, want EnvelopeAdvanced newCursor %s", cursor.Hex(), adv.NewCursor.Hex())
	}

	// ── setStatus(): complete the mandate ───────────────────────────────────
	tx, err := client.SetStatus(context.Background(), id, erc8312.StatusCompleted)
	if err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	receipt, err := bind.WaitMined(context.Background(), rpc, tx)
	if err != nil {
		t.Fatalf("WaitMined(setStatus %s): %v", tx.Hash().Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("setStatus reverted (tx %s)", tx.Hash().Hex())
	}
	status, err := client.GetStatus(context.Background(), id)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status != erc8312.StatusCompleted {
		t.Errorf("getStatus = %d, want %d (Completed)", status, erc8312.StatusCompleted)
	}
	active, err = client.IsActive(context.Background(), id)
	if err != nil {
		t.Fatalf("IsActive after complete: %v", err)
	}
	if active {
		t.Error("isActive = true after Completed, want false")
	}

	// ── unknown ids must error (the mock reverts) ───────────────────────────
	if _, err := client.GetEnvelope(context.Background(), common.Hash{}); err == nil {
		t.Error("GetEnvelope(zero) returned nil error, want revert for unknown id")
	}
}

func TestERC8312BudgetSubstrate(t *testing.T) {
	_, budgetAddr, _ := erc8312Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key := anvilSignerKey(t)
	signer := crypto.PubkeyToAddress(key.PublicKey)

	// The budget contract implements IBoundedAgentAction too — register the
	// envelope through the bounded client pointed at the same address.
	bounded := erc8312.NewBoundedAgentActionClient(rpc, budgetAddr, key)
	id, err := bounded.RegisterEnvelope(context.Background(), signer, crypto.Keccak256Hash([]byte("budget-test")), erc8312ExpiresAt, nil)
	if err != nil {
		t.Fatalf("RegisterEnvelope(budget): %v", err)
	}

	budget := erc8312.NewBudgetSubstrateClient(rpc, budgetAddr)

	// ── bound() / spent() / remaining(): the recompute-to-verify inputs ─────
	bound, err := budget.Bound(context.Background(), id)
	if err != nil {
		t.Fatalf("Bound: %v", err)
	}
	t.Logf("bound → cap=%d asset=%s", bound.Cap, bound.Asset.Hex())
	if bound.Cap != erc8312BudgetCap {
		t.Errorf("bound cap = %d, want %d (mock DEFAULT_CAP)", bound.Cap, erc8312BudgetCap)
	}
	if bound.Asset == (common.Address{}) {
		t.Error("bound asset = zero address, want USDC mock address")
	}
	spent, err := budget.Spent(context.Background(), id)
	if err != nil {
		t.Fatalf("Spent: %v", err)
	}
	if spent != 0 {
		t.Errorf("spent = %d at registration, want 0", spent)
	}
	remaining, err := budget.Remaining(context.Background(), id)
	if err != nil {
		t.Fatalf("Remaining: %v", err)
	}
	if remaining != erc8312BudgetCap {
		t.Errorf("remaining = %d at registration, want %d (full cap)", remaining, erc8312BudgetCap)
	}
	// Off-chain == on-chain: the untouched budget satisfies the invariant.
	if !erc8312.VerifyRemaining(bound.Cap, spent, remaining) {
		t.Error("VerifyRemaining(cap, 0, cap) = false on a fresh envelope")
	}

	// ── advanceCursor with a budget-substrate witness (abi.encode(uint256)) ─
	witness := packUint256(t, erc8312BudgetDelta)
	adv, err := bounded.AdvanceCursor(context.Background(), id, witness)
	if err != nil {
		t.Fatalf("AdvanceCursor(budget, delta=%d): %v", erc8312BudgetDelta, err)
	}
	spent, err = budget.Spent(context.Background(), id)
	if err != nil {
		t.Fatalf("Spent after advance: %v", err)
	}
	if spent != erc8312BudgetDelta {
		t.Errorf("spent = %d after advance, want %d", spent, erc8312BudgetDelta)
	}
	remaining, err = budget.Remaining(context.Background(), id)
	if err != nil {
		t.Fatalf("Remaining after advance: %v", err)
	}
	if remaining != erc8312BudgetCap-erc8312BudgetDelta {
		t.Errorf("remaining = %d after advance, want %d", remaining, erc8312BudgetCap-erc8312BudgetDelta)
	}
	// Off-chain == on-chain: recompute the headroom and the cursor root.
	if got := erc8312.ComputeRemainingHeadroom(bound.Cap, spent); got != remaining {
		t.Errorf("ComputeRemainingHeadroom(%d, %d) = %d, want on-chain remaining %d", bound.Cap, spent, got, remaining)
	}
	if !erc8312.VerifyRemaining(bound.Cap, spent, remaining) {
		t.Error("VerifyRemaining(cap, spent, remaining) = false after advance")
	}
	// The budget substrate pins cursorRoot = keccak256(abi.encode(spent)) —
	// the emitted newCursor must equal the off-chain recomputation.
	wantCursor := crypto.Keccak256Hash(packUint256(t, spent))
	if adv.NewCursor != wantCursor {
		t.Errorf("EnvelopeAdvanced newCursor = %s, want keccak256(abi.encode(spent)) %s", adv.NewCursor.Hex(), wantCursor.Hex())
	}

	// ── over-the-cap advance must revert (enforcement, not recompute) ───────
	overWitness := packUint256(t, erc8312BudgetCap) // spent 500 + 10000 > cap 10000
	if _, err := bounded.AdvanceCursor(context.Background(), id, overWitness); err == nil {
		t.Error("AdvanceCursor over the cap returned nil error, want revert (exceeds cap)")
	} else {
		t.Logf("over-cap advance rejected: %v", err)
	}
	// State must be unchanged after the revert.
	spent, err = budget.Spent(context.Background(), id)
	if err != nil {
		t.Fatalf("Spent after rejected advance: %v", err)
	}
	if spent != erc8312BudgetDelta {
		t.Errorf("spent = %d after rejected advance, want %d (unchanged)", spent, erc8312BudgetDelta)
	}
}

func TestERC8312ContestableEnvelope(t *testing.T) {
	_, _, contestableAddr := erc8312Addresses(t)

	rpc, err := ethclient.Dial(anvilRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial(%s): %v", anvilRPC, err)
	}
	defer rpc.Close()

	key := anvilSignerKey(t)
	signer := crypto.PubkeyToAddress(key.PublicKey)

	// The contestable contract implements IBoundedAgentAction too — register
	// through the bounded client pointed at the same address, then drive the
	// contestation lifecycle through the contestable client.
	bounded := erc8312.NewBoundedAgentActionClient(rpc, contestableAddr, key)
	id, err := bounded.RegisterEnvelope(context.Background(), signer, crypto.Keccak256Hash([]byte("contest-test")), erc8312ExpiresAt, nil)
	if err != nil {
		t.Fatalf("RegisterEnvelope(contestable): %v", err)
	}

	contestable := erc8312.NewContestableEnvelopeClient(rpc, contestableAddr, key)

	// ── contest(): Active -> Contested ──────────────────────────────────────
	info, err := contestable.Contest(context.Background(), id, []byte("evidence"))
	if err != nil {
		t.Fatalf("Contest: %v", err)
	}
	t.Logf("contest → challenger=%s", info.Challenger.Hex())
	if info.Challenger != signer {
		t.Errorf("EnvelopeContested challenger = %s, want signer %s", info.Challenger.Hex(), signer.Hex())
	}
	status, err := bounded.GetStatus(context.Background(), id)
	if err != nil {
		t.Fatalf("GetStatus after contest: %v", err)
	}
	if status != erc8312.StatusContested {
		t.Errorf("getStatus after contest = %d, want %d (Contested)", status, erc8312.StatusContested)
	}

	// ── resolve(): Contested -> Active ──────────────────────────────────────
	resolved, err := contestable.Resolve(context.Background(), id, erc8312.StatusActive, []byte("resolution"))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	t.Logf("resolve → outcome=%d", resolved.Outcome)
	if resolved.Outcome != erc8312.StatusActive {
		t.Errorf("EnvelopeResolved outcome = %d, want %d (Active)", resolved.Outcome, erc8312.StatusActive)
	}
	status, err = bounded.GetStatus(context.Background(), id)
	if err != nil {
		t.Fatalf("GetStatus after resolve: %v", err)
	}
	if status != erc8312.StatusActive {
		t.Errorf("getStatus after resolve = %d, want %d (Active)", status, erc8312.StatusActive)
	}
	active, err := bounded.IsActive(context.Background(), id)
	if err != nil {
		t.Fatalf("IsActive after resolve: %v", err)
	}
	if !active {
		t.Error("isActive = false after resolve-to-Active, want true")
	}
}

// erc8312Addresses reads the ERC8312_ADDRESSES env var set from deploy.sh
// output: three addresses, newline-separated in broadcast order
// (boundedAgentAction, budgetSubstrate, contestableEnvelope).
func erc8312Addresses(t *testing.T) (common.Address, common.Address, common.Address) {
	t.Helper()
	fields := strings.Fields(os.Getenv("ERC8312_ADDRESSES"))
	if len(fields) != 3 {
		t.Fatalf("ERC8312_ADDRESSES must hold 3 addresses (got %d) — deploy first via testkit/scripts/deploy.sh metering/ERC8312 DeployERC8312", len(fields))
	}
	return common.HexToAddress(fields[0]), common.HexToAddress(fields[1]), common.HexToAddress(fields[2])
}

// packUint256 abi-encodes a uint256 — the witness shape the budget substrate
// mock decodes (abi.decode(witness, (uint256))).
func packUint256(t *testing.T, v uint64) []byte {
	t.Helper()
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		t.Fatalf("abi.NewType(uint256): %v", err)
	}
	packed, err := abi.Arguments{{Type: uint256Type}}.Pack(new(big.Int).SetUint64(v))
	if err != nil {
		t.Fatalf("pack uint256 %d: %v", v, err)
	}
	return packed
}
