# ERC-8281 — Observation Commitment Protocol (Go)

**Recompute-to-verify: YES.**

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| `digest` is the commitment of observation `observation` (`digest = keccak256(observation)`) | **YES — pure recompute** | The digest is a deterministic hash of the observation bytes. Anyone holding the primary artifact can re-derive it without trusting a third party. |
| The commitment was recorded at the contract | **YES — contract-level verify** | The `Recorded` event log is the public ledger (the interface has no getter); `CheckRecorded` re-scans the contract's logs for the matching digest. |

## Layer 1 — Contract wrappers

**`ObservationCommitmentClient`** anchors and inspects commitments on a
deployed ERC-8281 contract.

| Method | Description |
|--------|-------------|
| `Record(digest common.Hash) (*types.Transaction, error)` | Send `record(digest)` — emits `Recorded(digest, committer)`. Returns the signed, broadcast transaction. Requires a signer key (chain id is fetched from the RPC at call time). Returns `ErrNoSigner` if the client has none. |
| `CheckRecorded(digest common.Hash) (bool, error)` | Scan the contract's `Recorded` event log for the digest from block 0. The event log is the ledger — no on-chain getter exists. |

The client is a concrete struct wrapping `*ethclient.Client`,
`common.Address`, and an optional signer key — no generics:

```go
import "github.com/trustless-ai/agent-sdk/go/verify/erc8281"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
key, _ := crypto.HexToECDSA("...")
client := erc8281.NewObservationCommitmentClient(rpc, common.HexToAddress(erc8281Address), key)
tx, err := client.Record(common.HexToHash("0x..."))
recorded, err := client.CheckRecorded(common.HexToHash("0x..."))
```

## Layer 2 — Pure recompute

One deterministic computation reproducible off-chain from public inputs:

- **`ComputeObservationDigest(observation []byte) common.Hash`** — `digest = keccak256(observation)` (ERC-8281 §1), the core OCP commitment step. Golden vector: `ComputeObservationDigest([]byte("hello"))` —> `0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8`. Keccak-256 never fails, so there is no error return; the function never panics.

The recompute function is pure — no RPC, no anvil, no deployed contract
required.

See `client.go` for the contract wrapper, `recompute.go` for the pure
function, and `go/test/erc8281_integration_test.go` for the testkit
integration test.
