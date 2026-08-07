# ERC-8203 — ConsultEscrow Settlement (Go)

**Recompute-to-verify: YES.**

## What IS recompute-to-verify: the settlement commitment

`release()` recomputes `commitmentHash = keccak256(abi.encode(jobId, resultHash))`
on-chain and only pays the provider if the EIP-191 personal_sign over that
commitment recovers to the job's attestor. Both `jobId` and `resultHash` are
public — `jobId` is a parameter, `resultHash` is posted as part of the
`release()` call and visible in the `Released` event. Anyone holding the
delivered result text can independently recompute the commitment and verify
it against the chain, without trusting a third party.

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| `verdictHash = keccak256(abi.encode(jobId, keccak256(utf8(resultText))))` | **YES — pure recompute** | Deterministic hash of public inputs (`jobId`, `resultText`). Recompute-kit recipe `8203/settlement-proof`. |
| The escrow actually released for `jobId` (`jobs(jobId).status == Released`) | **YES — contract-level verify** | Read-only `jobs(bytes32)` view call — the authoritative on-chain state, no gas, no key. |

## Layer 1 — Contract wrapper

**`ConsultEscrowClient`** wraps `IConsultEscrow` for a single deployed escrow
contract:

| Method | Description | State |
|--------|-------------|-------|
| `GetJob(ctx, jobId) (Job, error)` | Read the escrowed job (consumer, provider, attestor, amount, deadline, status) via the `jobs` view. A never-opened job reads as the all-zero default. | read-only |
| `Resolve(ctx, jobId, resultHash, signature) (*types.Transaction, error)` | Broadcast `release(jobId, resultHash, signature)` — pays the provider once the attestor's signature over the commitment is valid. Returns `ErrNoSigner` if the client has no key. | broadcast |

The client is a concrete struct wrapping `*ethclient.Client`,
`common.Address` and an optional signer key — no generics:

```go
import "github.com/trustless-ai/agent-sdk/go/settlement/erc8203"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
client := erc8203.NewConsultEscrowClient(rpc, common.HexToAddress(escrowAddress), key)
job, err := client.GetJob(context.Background(), jobID)
tx, err := client.Resolve(context.Background(), jobID, resultHash, signature)
```

The optional `*ecdsa.PrivateKey` (nil for read-only clients) signs the
`release()` transaction — following the ERC-8281 Go client precedent. The
attestor's protocol signature (65-byte `[r||s||v]`, `v = 27/28`) is a
separate argument; `Resolve` never signs for the attestor.

## Layer 2 — Pure recompute

One deterministic computation reproducible off-chain from public inputs:

- **`ComputeVerdictHash(jobId common.Hash, resultText string) (common.Hash, error)`** —
  the release commitment, `verdictHash = keccak256(abi.encode(bytes32 jobId, keccak256(utf8(resultText))))`.
  Golden vector (recompute-kit `8203/settlement-proof`, cross-verified
  against the TypeScript, Python and Rust SDKs):
  - `jobId: 0xbc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56`
  - `resultText: "No intermediaries required, cryptographic verification only."`
  - expected: `0xdc568bd1cbacdd1ead8231e9d3d6f4e475f5168f3cc9f72b31935d46cfdd48f7`

The recompute function is pure — no RPC, no anvil, no deployed contract
required. Keccak-256 itself never fails; the `error` return covers the ABI
type construction and packing, and the function never panics.

See `client.go` for the contract wrapper, `recompute.go` for the pure
function, and `go/test/erc8203_integration_test.go` for the testkit
integration test (open a job, recompute the commitment, sign it as the
attestor, release, and check the on-chain commitment matches the recompute).
