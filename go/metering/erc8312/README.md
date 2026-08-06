# ERC-8312 — Bounded Agent Actions (Go)

**Recompute-to-verify: YES (four functions).**

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| `(reserved + confirmed) <= cap` (StatefulBound variant) | **YES — pure recompute** | Pure integer invariant over public quantities. |
| `aggregate <= cap` (Orbmis/headroom variant) | **YES — pure recompute** | Pure integer invariant over public quantities. |
| `remaining = cap - spent` (IBudgetSubstrate profile) | **YES — pure recompute** | The budget contract exposes `bound`, `spent`, `remaining` as views, so anyone can recompute headroom off-chain and cross-check the on-chain value. |
| Cursor advancement respected the bound (no bypass) | **CONTRACT-LEVEL** | Enforcement is a substrate property (`advanceCursor` reverts when the witness would exceed the cap) — observable, not recomputable. |

## Layer 2 — Pure recompute

Four deterministic computations reproducible off-chain from public integers —
no RPC, no anvil, no deployed contract:

- **`CheckStatefulBound(reserved, confirmed, cap uint64) bool`** —
  `(reserved + confirmed) <= cap`, computed without overflow. Golden vector
  (recompute-kit `8312/cap-conservation`): `(100, 0, 150)` → `true`;
  breach `(100, 60, 150)` → `false`.
- **`CheckCursorHeadroom(aggregate, cap uint64) bool`** — `aggregate <= cap`.
  Golden vector: `(0, 8000)` → `true`; breach `(8001, 8000)` → `false`.
- **`ComputeRemainingHeadroom(cap, spent uint64) uint64`** —
  `remaining = cap - spent` (ERC-8312 §IBudgetSubstrate), saturating:
  returns `0` when `spent > cap` (exhausted or inactive envelope). Golden
  vector: `(150, 60)` → `90`; `(150, 200)` → `0`.
- **`VerifyRemaining(cap, spent, reported uint64) bool`** — `spent <= cap`
  AND `(cap - spent) == reported`. `remaining(id)` is recomputed, never
  trusted. **CRITICAL:** the `spent <= cap` guard is load-bearing — the
  saturating subtraction makes `cap - spent == 0` when `spent > cap`, so
  without the guard a spent-over-cap envelope would pass with `reported=0`.

The vectors come from recompute-kit conformance (`8312/cap-conservation`,
`8312/budget-substrate`) and are cross-verified against the TypeScript,
Python and Rust SDKs. The functions never panic and never fail — no error
returns.

## Layer 1 — Contract wrappers

Three typed clients, one per ERC-8312 interface. All use the established
`(rpc *ethclient.Client, address common.Address[, key *ecdsa.PrivateKey])`
constructor shape — `key` is required for write methods and `nil` for
read-only clients:

**`BoundedAgentActionClient`** (`IBoundedAgentAction` — envelope lifecycle):

| Method | Description |
|--------|-------------|
| `RegisterEnvelope(ctx, principal common.Address, capabilityRoot common.Hash, expiresAt uint64, initData []byte) (common.Hash, error)` | Broadcasts `registerEnvelope`, returns the contract-generated `id` from the `EnvelopeRegistered` event. |
| `AdvanceCursor(ctx, id common.Hash, witness []byte) (AdvanceResult, error)` | Broadcasts `advanceCursor`; returns prev/new cursor from the `EnvelopeAdvanced` event. |
| `SetStatus(ctx, id common.Hash, newStatus Status) (*types.Transaction, error)` | Broadcasts `setStatus(id, newStatus)`; wait with `bind.WaitMined`. |
| `GetEnvelope(ctx, id common.Hash) (Envelope, error)` | Reads the full stored envelope — the recompute-to-verify input. |
| `GetCursor(ctx, id common.Hash) (common.Hash, error)` | Reads the cursor commitment. |
| `GetStatus(ctx, id common.Hash) (Status, error)` | Reads the effective status (Expired once `expiresAt` is reached). |
| `IsActive(ctx, id common.Hash) (bool, error)` | True iff the effective status is Active. |

**`BudgetSubstrateClient`** (`IBudgetSubstrate` — budget profile, read-only):

| Method | Description |
|--------|-------------|
| `Bound(ctx, id common.Hash) (Bound, error)` | Reads `(cap, asset)`. |
| `Spent(ctx, id common.Hash) (uint64, error)` | Reads cumulative consumed value. |
| `Remaining(ctx, id common.Hash) (uint64, error)` | Reads `cap - spent` (0 when inactive) — cross-check with `VerifyRemaining`. |

**`ContestableEnvelopeClient`** (`IContestableEnvelope` — contestation):

| Method | Description |
|--------|-------------|
| `Contest(ctx, id common.Hash, evidence []byte) (ContestInfo, error)` | Broadcasts `contest`; returns the challenger from the `EnvelopeContested` event. |
| `Resolve(ctx, id common.Hash, outcome Status, resolution []byte) (ResolveInfo, error)` | Broadcasts `resolve`; returns the outcome from the `EnvelopeResolved` event. |

```go
import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trustless-ai/agent-sdk/go/metering/erc8312"
)

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
bounded := erc8312.NewBoundedAgentActionClient(rpc, common.HexToAddress(boundedAddr), key)
id, _ := bounded.RegisterEnvelope(ctx, signer, capabilityRoot, 2000000000, nil)

budget := erc8312.NewBudgetSubstrateClient(rpc, common.HexToAddress(budgetAddr))
bound, _ := budget.Bound(ctx, id)
spent, _ := budget.Spent(ctx, id)
remaining, _ := budget.Remaining(ctx, id)
if !erc8312.VerifyRemaining(bound.Cap, spent, remaining) {
	// on-chain headroom diverges from the recomputed invariant
}
```

The clients are concrete structs wrapping `*ethclient.Client`,
`common.Address` and an optional signer key — no generics, no panics (every
failure is a returned error).

## Integration test

`go/test/erc8312_integration_test.go` deploys all three ERC-8312 mocks via
`testkit/scripts/deploy.sh metering/ERC8312 DeployERC8312` — three addresses
in `ERC8312_ADDRESSES`, newline-separated in broadcast order
(boundedAgentAction, budgetSubstrate, contestableEnvelope):

```bash
testkit/scripts/start-anvil.sh
ERC8312_ADDRESSES=$(testkit/scripts/deploy.sh metering/ERC8312 DeployERC8312)
ERC8312_ADDRESSES="$ERC8312_ADDRESSES" go test -v ./go/test/ -run TestERC8312
testkit/scripts/stop-anvil.sh
```

The test drives the full lifecycle on each contract and cross-checks the
on-chain budget state against the recompute functions
(`ComputeRemainingHeadroom` / `VerifyRemaining`), including an
over-the-cap `advanceCursor` that the mock must revert.

See `client.go` for the contract wrappers, `recompute.go` for the pure
functions, and `go/metering/erc8312/recompute_test.go` for the conformance
vectors.
