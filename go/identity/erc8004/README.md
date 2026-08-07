# ERC-8004 — Agent Identity Registry (Go)

**Recompute-to-verify: SPLIT (one claim YES, one claim NO).**

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| `agentId = bytes32(uint256(registryId))` | **YES — pure recompute** | The registry assigns every agent an integer id; the on-chain agentId is that id left-padded to 32 bytes — not a hash. Anyone who knows the registry id can independently recompute the agentId without trusting a third party. |
| Registration state (URI, metadata, payment wallet) | **NOT recomputable** | Written by the agent owner; there is no derivation rule that reproduces it off-chain. The registry contract is the source of truth — read it directly with a read-only `eth_call` (no gas, no key). |

## Layer 1 — Contract wrappers

**`IdentityRegistryClient`** reads identity state from a deployed ERC-8004 contract.

| Method | Description |
|--------|-------------|
| `GetAgentURI(tokenID)` | Read the registration-file URI (ERC-721 `tokenURI`). |
| `GetMetadata(agentID, key)` | Read an on-chain metadata value for an agent. |
| `GetAgentWallet(agentID)` | Read the agent's current payment wallet. |
| `OwnerOf(tokenID)` | Read the ERC-721 owner of an agent id. |

All calls are read-only view functions — no gas or broadcast needed. The
client is a concrete struct wrapping `*ethclient.Client` and
`common.Address` — no generics, no signing required for reads:

```go
import "github.com/trustless-ai/agent-sdk/go/identity/erc8004"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
client := erc8004.NewIdentityRegistryClient(rpc, common.HexToAddress(erc8004Address))
uri, err := client.GetAgentURI(context.Background(), big.NewInt(1))
```

## Layer 2 — Pure recompute

One deterministic transformation reproducible off-chain from public inputs:

- **`ComputeAgentId(registryId uint64) common.Hash`** — `agentId = bytes32(uint256(registryId))`, a left-padded zero-extension (not a hash). Golden vector: `registryId: 860` → `0x000000000000000000000000000000000000000000000000000000000000035c`. Never panics and cannot fail.

The recompute function is pure integer logic with inline golden vectors —
no RPC, no anvil, no deployed contract required.

See `client.go` for the contract wrapper, `recompute.go` for the pure
function, and `go/test/erc8004_integration_test.go` for the testkit
integration test.
