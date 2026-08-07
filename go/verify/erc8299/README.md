# ERC-8299 — WYRIWE: Input Provenance for AI Inference (Go)

**Recompute-to-verify: YES.**

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| `rawInputHash` commits to the raw user input (`rawInputHash = keccak256(raw_user_input)`, §45) | **YES — pure recompute** | A deterministic hash of the input bytes. Anyone holding the raw input can re-derive it without trusting a third party. |
| `sanitizationPipelineHash` commits to the pinned sanitization spec (`keccak256(utf8(cid) || rawInputHash)`, §46) | **YES — pure recompute** | A deterministic hash of the CID bytes and the recomputed `rawInputHash`. Fully reproducible from public data. |
| `rawProposalHash` commits to the proposed-action artifact (L4, `sha256(artifact)`) | **YES — pure recompute** | sha256 over the artifact's UTF-8 bytes — NOT keccak256, since L4 anchors off-chain (Nostr-relay-published) verdicts too. |
| `verdictHash` binds the verdict metadata to the proposal (L4, `sha256(JCS({preimage fields}))`) | **YES — pure recompute** | JCS (RFC 8785) canonical JSON of the producer's published `decision_ref_preimage_fields`, sha256'd — the exact `decision_ref` format. Null-valued preimage fields serialize as JSON `null`. |
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

Four deterministic computations reproducible off-chain from public inputs.
L1-L3 are keccak256-based; the L4 layer is sha256-based — see the doc
comments in `recompute.go` for why.

**L1-L3 (keccak256):**

- **`ComputeRawInputHash(rawInput []byte) common.Hash`** — `rawInputHash = keccak256(raw_user_input)` (ERC-8299 §45). Golden vector: `ComputeRawInputHash([]byte("hello"))` —> `0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8`.
- **`ComputeSanitizationPipelineHash(cid string, rawInputHash common.Hash) common.Hash`** — `sanitizationPipelineHash = keccak256(utf8(cid) || rawInputHash)` (ERC-8299 §46): the CID UTF-8 bytes concatenated with the raw-input-hash bytes, then hashed. Golden vector: `ComputeSanitizationPipelineHash("ipfs://QmTest", 0x1c8a…deac8)` —> `0xb9b571ee6d24c3fcd09fcca0811099b00920d274d0a4b2612531201b8a6f35c1` (recompute-kit conformance `wyriwe/raw`, `wyriwe/pipeline` also cross-verified against the TS, Python and Rust SDKs).

**L4 (sha256, cross-lane vectors in `testkit/vectors/erc8299-l4.vectors.json`):**

- **`ComputeRawProposalHash(artifact string) string`** — `rawProposalHash = sha256(artifact)`: sha256 over the artifact's UTF-8 bytes (NOT keccak256 — L4 also anchors off-chain verdicts, matching invinoveritas's reference implementation). Returns `0x`-prefixed hex. Golden vectors: `ComputeRawProposalHash("test artifact content for cross-language verification")` —> `0xb8f70a237da212a272ecd09370acedbce6ca1d7df90745beafcac77e39697a88`; empty artifact —> `0xe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` (sha256 of empty, a legal input).
- **`ComputeVerdictHash(fields map[string]string, preimageFields []string) string`** — `verdictHash = sha256(JCS({preimage fields}))`: JCS (RFC 8785) canonical JSON of the preimage fields — keys sorted by code point (Go `sort.Strings` = UTF-8 byte order = code-point order), each key/value JSON-encoded with no extraneous whitespace and literal UTF-8, and keys missing from `fields` serialized as JSON `null` (never `""`). Returns `"sha256:" + hex`, matching invinoveritas's `decision_ref` format. Golden vectors: the real `/ledger` entry #236 —> `sha256:5bca0bf044c8e1c8e16a01bf3ee44b12c305ce6a50dd9789ff73cbd13482b9b9`; null-valued field —> `sha256:2970854c035d5aedb673b8523128665712895f62dd525c91fc8e858ad588ce58`; non-ASCII key sort (a, Ｚ, 😀) —> `sha256:36e2e43ff6d7062ebb64c209604b7ce028b4eb88d4db2892e872194d16f36bca`.

The recompute functions are pure — no RPC, no anvil, no deployed contract
required. `recompute_test.go` runs the named goldens plus every vector from
`testkit/vectors/erc8299-l4.vectors.json` programmatically.

See `client.go` for the contract wrappers, `recompute.go` for the pure
functions, and `go/test/erc8299_integration_test.go` for the testkit
integration test (deploys both ERC-8299 mocks via
`testkit/scripts/deploy.sh verify/ERC8299 DeployERC8299` — two addresses in
`ERC8299_ADDRESSES`, newline-separated in broadcast order).
