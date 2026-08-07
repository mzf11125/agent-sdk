# ERC-8323 — Source-Token Agent Binding (Go)

**Recompute-to-verify: NO.**

ERC-8323 binds an ERC-8004 agent identity to a pre-existing ERC-721 token in
a single bound collection. The contract records `(boundCollection,
sourceTokenId)` as immutable provenance at registration (exactly one
`SourceNFTLinked` event per agent) and exposes live ownership separately,
re-checked at query time. Source binding is an **on-chain fact** — the
registry is the source of truth, and there is no deterministic derivation
reproducible off-chain from public inputs (unlike the hash-based ERCs 8299,
8301, 8203, ...). Verification is a direct `getSourceNFT` /
`isSourceNFTOwnershipValid` read, not a hash re-derivation, so no recompute
function is generated here.

## Layer 1 — Contract wrappers

**`SourceBindingClient`** reads source-binding state from and registers agents
on a deployed ERC-8323 contract (`IAgentSourceBinding`, ERC-165 interface id
`0x27eba962`).

| Method | Description |
|--------|-------------|
| `BoundCollection(ctx) (common.Address, error)` | Read the source ERC-721 collection this registry is bound to (immutable). |
| `Register(ctx, sourceTokenID, value *big.Int) (*big.Int, error)` | Broadcast `registerWithSource(sourceTokenID)` with `value` wei as `msg.value`, wait for the receipt, and return the minted agent id from `SourceNFTLinked`. Reverts on insufficient value — pass the registry's mint price (wei) if it charges one. Requires a signer key; returns `ErrNoSigner` if the client has none. |
| `GetSourceNFT(ctx, agentID) (SourceNFT, error)` | Read the immutable `(sourceContract, sourceTokenId)` an agent was derived from. Errors for non-existent / unbound agents. |
| `HasSourceNFT(ctx, agentID) (bool, error)` | Whether the agent has a recorded source binding. |
| `IsSourceNFTOwnershipValid(ctx, agentID) (bool, error)` | Whether the source token is still under the agent's control — the live 3-case check (direct owner, canonical ERC-6551 TBA, or the binding contract itself), rechecked at query time. |
| `OwnerOf(ctx, agentID) (common.Address, error)` | Read the ERC-721 owner of an agent id (a compliant registry MUST also implement ERC-721). |
| `SupportsSourceBinding(ctx) (bool, error)` | ERC-165 check for `IAgentSourceBinding` (`0x27eba962`). |

`SourceBindingInterfaceID` (`0x27eba962`) is exported for direct ERC-165
checks. The client is a concrete struct wrapping `*ethclient.Client`,
`common.Address`, and an optional signer key — no generics:

```go
import "github.com/trustless-ai/agent-sdk/go/identity/erc8323"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
key, _ := crypto.HexToECDSA("...")
client := erc8323.NewSourceBindingClient(rpc, common.HexToAddress(erc8323Address), key)

agentID, err := client.Register(context.Background(), big.NewInt(42), big.NewInt(1e15)) // 0.001 ether mint price
src, err := client.GetSourceNFT(context.Background(), agentID)                            // (collection, 42)
valid, err := client.IsSourceNFTOwnershipValid(context.Background(), agentID)
```

Read-only methods need no key — pass `nil` to `NewSourceBindingClient` for a
read-only client.

## Layer 2 — Pure recompute: NONE

The ERC defines no deterministic computation reproducible off-chain from
public inputs, so there are no conformance vectors and no recompute module.

See `client.go` for the contract wrapper, `types.go` for the `SourceNFT` and
`SourceNFTLinkedEvent` types, and `go/test/erc8323_integration_test.go` for
the testkit integration test.
