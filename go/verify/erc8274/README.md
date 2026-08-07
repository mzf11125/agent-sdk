# ERC-8274 — AI Inference Proof Verification (Go)

**Recompute-to-verify: NO.**

ERC-8274 is three separate interfaces meant to be deployed as three separate
contracts that reference each other by address:

```
Settlement Contract (IAgentVerifiable) — declares which IAgentVerifier it trusts
    └── IAgentVerifier — stateful: agent authorization + proof routing
            └── IProofVerifier — stateless: raw cryptographic proof check (zkML/opML/TEE)
```

The proof-validity result itself is checked on-chain by the deployed,
immutable `IProofVerifier` contract, which anyone can call directly —
`ProofVerifierClient.Verify()` is that call, exposed as a free read-only
`eth_call` (the same trust model as re-checking a Merkle proof yourself:
you trust the public, auditable checking procedure once, not each party who
invokes it). This package performs **no off-chain recomputation** of its
own: the SDK deliberately delegates the check rather than reimplementing
proof-system cryptography, and the `verificationDigest` from
`IAgentVerifier.verify()` cannot be reconstructed from the standard
interface alone anyway — one of its six preimage components
(`agentProofProfile`) is not exposed by `IAgentVerifier` (see
`agent-ercs/verify/ERC8274`).

## Layer 1 — Contract wrappers

**`ProofVerifierClient`** wraps `IProofVerifier` (read-only — no gas, no key):

| Method | Description |
|--------|-------------|
| `ProofSystem() (string, error)` | Read the `proofSystem()` view — human-readable proof-system identifier (e.g. "zkML-Halo2"). |
| `ProofProfile() (common.Hash, error)` | Read the `proofProfile()` view — compact proof profile hash. |
| `Verify(inputHash, outputHash common.Hash, metadata []byte, proof common.Hash) (bool, error)` | `eth_call` of `verify(...)` — the boolean is computed on-chain and never trusted off-chain. |

**`AgentVerifierClient`** wraps `IAgentVerifier` (broadcasts — records the
`VerificationCompleted` event on-chain, which is the point of calling it):

| Method | Description |
|--------|-------------|
| `VerifyTask(taskId, agentId, inputHash, outputHash common.Hash) (bool, error)` | Derives the proof via `GetDigest`, broadcasts `verify(...)`, waits for the receipt and returns the `valid` flag parsed from the `VerificationCompleted` event. Requires a signer key; returns `ErrNoSigner` without one. |
| `GetDigest(inputHash, outputHash common.Hash) (common.Hash, error)` | Pure off-chain computation: `keccak256(abi.encodePacked(inputHash, outputHash))` — the expected proof digest of the empty-metadata routing (the ERC doesn't specify per-agent metadata; the testkit mock routes with empty metadata). Always returns a nil error. |

**`GetTrustedVerifier(rpc, verifiableAddr common.Address) (common.Address, error)`**
— a standalone function, not a client struct: `IAgentVerifiable` is a single
getter, so a struct would be one method wrapping a constructor for no
benefit. Reads the `agentVerifier()` address a settlement contract trusts.

The clients are concrete structs wrapping `*ethclient.Client`,
`common.Address`, and an optional signer key — no generics:

```go
import "github.com/trustless-ai/agent-sdk/go/verify/erc8274"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")

proofVerifier := erc8274.NewProofVerifierClient(rpc, common.HexToAddress(proofVerifierAddr))
valid, err := proofVerifier.Verify(inputHash, outputHash, metadata, proof)

key, _ := crypto.HexToECDSA("...")
agentVerifier := erc8274.NewAgentVerifierClient(rpc, common.HexToAddress(agentVerifierAddr), key)
valid, err := agentVerifier.VerifyTask(taskId, agentId, inputHash, outputHash)

trusted, err := erc8274.GetTrustedVerifier(rpc, common.HexToAddress(verifiableAddr))
```

## Layer 2 — Pure recompute: NONE

The ERC defines no deterministic computation reproducible off-chain from
public inputs that this SDK could own: proof validity is delegated to the
deployed contract (that call IS the recompute mechanism), and the
`verificationDigest` is missing one preimage component from the standard
interface. `GetDigest` is a mock-alignment convenience, not a recompute
function.

See `client.go` for the contract wrappers and `go/test/erc8274_integration_test.go`
for the testkit integration test (three addresses from
`testkit/scripts/deploy.sh verify/ERC8274 DeployERC8274`).
