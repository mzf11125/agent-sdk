# ERC-8275 — Agent Reputation (Go)

**Recompute-to-verify: SPLIT (one claim YES, one claim YES-as-contract-call, one claim NO).**

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| Win rate (`winRate = wins/(wins+losses)`) | **YES — pure recompute** | Deterministic arithmetic from public inputs (wins/losses). Anyone with the event log can independently recompute this without trusting a third party. |
| Composite score (`f(attestationCount, counterpartyDiversity, winRate, volumeCap)`) | **NOT verifiable** | The spec defines only a "recommended scoring shape" — the exact function is implementation-defined, so a generic SDK cannot reproduce it. |
| `verifyOutcome` (on-chain proof check) | **YES — contract-level verify** | A `view` function anyone can call via a read-only `eth_call` (no gas, no key). Gives the contract's authoritative answer without broadcasting a transaction. |

## Layer 1 — Contract wrappers

**`AgentReputationClient`** reads reputation state from a deployed ERC-8275 contract.

| Method | Description |
|--------|-------------|
| `GetReputation(agentID)` | Read the current reputation snapshot (CompletedOrders, DisputedOrders, TotalVolume, LastActiveAt, Score). |
| `GetDecayWeight(agentID)` | Read the recency-decay weight in basis points. |
| `VerifyOutcome(orderID, proof)` | Read-only `eth_call` — verify a settled order's outcome proof against the public record. |

All three calls are read-only; no gas or broadcast needed. The client is a
concrete struct wrapping `*ethclient.Client` and `common.Address` — no
generics, no signing required for reads:

```go
import "github.com/trustless-ai/agent-sdk/go/reputation/erc8275"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
client := erc8275.NewAgentReputationClient(rpc, common.HexToAddress(erc8275Address))
rep, err := client.GetReputation(context.Background(), common.Hash{})
```

## Layer 2 — Pure recompute

One deterministic computation reproducible off-chain from public inputs:

- **`ComputeWinRate(wins, losses uint64) (uint64, error)`** — `winRate = gated_wins / (gated_wins + gated_losses)`, rounded to 4 decimal places (ERC-8275). Golden vector: `wins: 16, losses: 15` —> `5161` (0.5161). Returns `ErrZeroTotal` when both wins and losses are zero. Never panics.

Convention binding (`win_rate.bps.v0`, content-addressed per recompute-kit
convention-hash-v0): `PinWinRateBps` stamps the governing convention hash at
issuance; `VerifyWinRate` resolves at verification — tri-state, fail-closed
(`Verified` / `Rejected` / `Unverifiable`).

The recompute functions are pure integer arithmetic with inline golden
vectors — no RPC, no anvil, no deployed contract required.

See `client.go` for the contract wrapper, `recompute.go` for the pure
function, `conventions.go` for the convention binding, and
`go/test/erc8275_integration_test.go` for the testkit integration test.
