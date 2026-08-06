# ERC-8299 — WYRIWE: Input Provenance for AI Inference (Go)

**Recompute-to-verify: YES.**

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| `rawInputHash` commits to the raw user input (`rawInputHash = keccak256(raw_user_input)`, §45) | **YES — pure recompute** | A deterministic hash of the input bytes. Anyone holding the raw input can re-derive it without trusting a third party. |
| `sanitizationPipelineHash` commits to the pinned sanitization spec (`keccak256(utf8(cid) || rawInputHash)`, §46) | **YES — pure recompute** | A deterministic hash of the CID bytes and the recomputed `rawInputHash`. Fully reproducible from public data. |
| The attestation's EIP-712-style signature is authentic (`verify(attestation, signature)`) | **CONTRACT-LEVEL verify** | The signature is checked by the contract against the known attestor — a contract read, not a recompute. The L3/L4 client exposes it as a read-only view call (no gas, no signer). |

## Layer 1 — Contract wrappers

Two clients, one per ERC-8299 contract layer. Both are read-only — no gas or
funded key needed, which is the whole point of exposing verification this
way.

**`WyriweAttestationClient`** (L3 — input provenance,
`IWyriweAttestation`):

| Method | Description |
|--------|-------------|
| `Verify(ctx, attestation WyriweAttestation, signature []byte) (bool, error)` | Read-only `verify(attestation, signature)` view call against the known attestor. |
| `ProofSystem(ctx) (string, error)` | Reads `proofSystem()` — always `"attestation/wyriwe"` for conforming implementations. |

**`JudgmentExecutionClient`** (L4 — judgment validator chain-of-custody,
`IJudgmentExecutionAttestation`):

| Method | Description |
|--------|-------------|
| `Verify(ctx, attestation JudgmentExecutionAttestation, signature []byte) (bool, error)` | Read-only `verify(attestation, signature)` view call against the executing-agent attestor. |
| `ProofSystem(ctx) (string, error)` | Reads `proofSystem()` — always `"attestation/judgment"`. |

The clients are concrete structs wrapping `*ethclient.Client` and
`common.Address` — no generics. The tuple struct arguments
(`WyriweAttestation`, `JudgmentExecutionAttestation`) are packed by
go-ethereum from the exported struct fields, which mirror the ABI component
names:

```go
import "github.com/trustless-ai/agent-sdk/go/verify/erc8299"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
client := erc8299.NewWyriweAttestationClient(rpc, common.HexToAddress(wyriweAddress))
valid, err := client.Verify(ctx, attestation, signature)
system, err := client.ProofSystem(ctx)
```

## Layer 2 — Pure recompute

Two deterministic computations reproducible off-chain from public inputs:

- **`ComputeRawInputHash(rawInput []byte) common.Hash`** — `rawInputHash = keccak256(raw_user_input)` (ERC-8299 §45). Golden vector: `ComputeRawInputHash([]byte("hello"))` —> `0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8`.
- **`ComputeSanitizationPipelineHash(cid string, rawInputHash common.Hash) common.Hash`** — `sanitizationPipelineHash = keccak256(utf8(cid) || rawInputHash)` (ERC-8299 §46): the CID UTF-8 bytes concatenated with the raw-input-hash bytes, then hashed. Golden vector: `ComputeSanitizationPipelineHash("ipfs://QmccvoM6aRVgZ2dtFWvT6Wm3DmTvoAUHHotK7uQufnStVR", 0x1c8a…deac8)` —> `0x5798efed4aa92f96a0622fc30268042b067294bdb5fd06f599bf8d84fd5d734b`.

Both vectors come from recompute-kit conformance (`wyriwe/raw`,
`wyriwe/pipeline`) and are cross-verified against the TypeScript, Python and
Rust SDKs. Keccak-256 never fails, so there are no error returns; the
functions never panic.

The recompute functions are pure — no RPC, no anvil, no deployed contract
required.

See `client.go` for the contract wrappers, `recompute.go` for the pure
functions, and `go/test/erc8299_integration_test.go` for the testkit
integration test (deploys both ERC-8299 mocks via
`testkit/scripts/deploy.sh verify/ERC8299 DeployERC8299` — two addresses in
`ERC8299_ADDRESSES`, newline-separated in broadcast order).
