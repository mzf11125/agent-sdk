# ERC-8263 — OnChain Proof Anchor (Go)

**Recompute-to-verify: NO.**

ERC-8263 is a write-side anchor floor. The contract emits `AnchorProof`
events with a non-zero `proofHash` commitment together with an
identity-scheme byte and a 32-byte `agentId`. It performs no verification and
no profile detection — `proofHash` is explicitly "an opaque non-zero bytes32
commitment". Since the ERC does **not** fix how `proofHash` is computed, a
generic SDK cannot independently recompute it. Verification logic exists at
the profile/application layer, not the ERC layer, so no recompute function is
generated here.

## Layer 1 — Contract wrappers

**`OnChainProofClient`** anchors proofs on a deployed ERC-8263 contract.

| Method | Description |
|--------|-------------|
| `Anchor(agentIdScheme uint8, agentId, proofHash common.Hash) (*types.Transaction, error)` | Send `anchor(agentIdScheme, agentId, proofHash)` — emits `AnchorProof` with empty `aux`. Returns the signed, broadcast transaction. Requires a signer key (chain id is fetched from the RPC at call time). Returns `ErrNoSigner` if the client has none. |
| `AnchorWithAux(agentIdScheme uint8, agentId, proofHash common.Hash, aux []byte) (*types.Transaction, error)` | Send `anchorWithAux(...)` with opaque extension bytes (non-normative). Same guards and semantics as `Anchor`. |
| `IsAnchored(proofHash common.Hash) (bool, error)` | Scan the contract's `AnchorProof` event log for the proof hash from block 0. The event log is the ledger — no on-chain getter exists. |

The client is a concrete struct wrapping `*ethclient.Client`,
`common.Address`, and an optional signer key — no generics:

```go
import "github.com/trustless-ai/agent-sdk/go/anchor/erc8263"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
key, _ := crypto.HexToECDSA("...")
client := erc8263.NewOnChainProofClient(rpc, common.HexToAddress(erc8263Address), key)
tx, err := client.Anchor(0x01, agentID, proofHash)
anchored, err := client.IsAnchored(proofHash)
```

### Canonical-form guards

The contract (and the `testkit` mock used by the integration tests) enforces
write-time invariants, surfaced as revert errors during gas estimation:

- `proofHash` MUST be non-zero.
- Scheme `0x00` (ANONYMOUS) requires `agentId == 0`.
- Schemes `0x01` (REGISTRY) / `0x02` (URI_HASH) require `agentId != 0`.
- Schemes `0x03+` are reserved and revert.

## Layer 2 — Pure recompute: NONE

The ERC defines no deterministic computation reproducible off-chain from
public inputs, and no conformance vectors exist in `recompute-kit` for
ERC-8263.

See `client.go` for the contract wrapper, `types.go` for the
`AnchorProofEvent` type, and `go/test/erc8263_integration_test.go` for the
testkit integration test.
