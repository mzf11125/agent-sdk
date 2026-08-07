# ERC-8301 — Agent Workflow / AgentTask (Go)

**Recompute-to-verify: SPLIT.**

| Claim | Verdict | Rationale |
|-------|---------|-----------|
| taskHash (`keccak256(abi.encode(stage, taskSeq, inputHash, timestamp, expiresAt, innerHash, workflowRunId))`) | **YES — pure recompute** | Deterministic function of public struct fields. The contract exposes the full struct via `getAgentTask()`, so anyone can independently recompute the hash without trusting the party who submitted the task. |
| replyHash (`keccak256(abi.encode(outputHash, timestamp, replier, innerHash, workflowRunId))`) | **YES — pure recompute** | Same shape as taskHash: pure function of the struct fields exposed by `getAgentReply()`. |
| Workflow FSM correctness (gate validation, stage progression) | **NOT verifiable** | The FSM transition logic is implementation-defined with no interface-level specification — no generic recompute check exists. |
| `getAgentTask` / `getAgentReply` (on-chain struct reads) | **YES — contract-level verify** | `view` functions anyone can call via a read-only `eth_call` (no gas, no key). Give the contract's authoritative answer without broadcasting. |

**Critical edge case:** when the prev-hash list is empty, the inner hash is
`keccak256(abi.encodePacked([])) = keccak256("") =
0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470` — NOT
`bytes32(0)`. `ComputeInnerHash` (and both hash functions) handle this
correctly; a naive zero-special-case silently breaks the evidence chain.

## Layer 1 — Contract wrappers

**`AgentWorkflowClient`** reads and drives workflow state on a deployed
ERC-8301 `IAgentWorkflow` contract. `GetTask`, `GetReply` and `Result` are
read-only `eth_call`s (no gas, no key); `Run`, `OnAgentReply` and
`OnAgentProve` broadcast transactions and need a signer key (nil for a
read-only client):

| Method | Description |
|--------|-------------|
| `Run(inputHash, input, expiresAt)` | Broadcast `run()`, wait for the receipt, parse the `NewAgentTask` event → `RunInfo{WorkflowRunId, TaskHash, Stage}`. |
| `Result(workflowRunId)` | Read the run's terminal state (`RunStatus`, final task hash, completion time). |
| `GetTask(taskHash)` | Read the stored `AgentTask` + proven status. Errors on unknown taskHash. |
| `GetReply(replyHash)` | Read the stored `AgentReply` + verifier, proven, verificationDigest. Errors on unknown replyHash. |
| `OnAgentReply(reply)` | Broadcast `onAgentReply(reply)` — the contract anchors the reply under its derived replyHash. `reply.Replier` MUST equal the signer. |
| `OnAgentProve(replyHashes, proof)` | Broadcast `onAgentProve(replyHashes, proof)` to mark anchored replies proven. |

The client is a concrete struct wrapping `*ethclient.Client`,
`common.Address` and an optional `*ecdsa.PrivateKey` — no generics:

```go
import "github.com/trustless-ai/agent-sdk/go/execution/erc8301"

rpc, _ := ethclient.Dial("http://127.0.0.1:8545")
client := erc8301.NewAgentWorkflowClient(rpc, common.HexToAddress(erc8301Address), signerKey)
info, err := client.Run(context.Background(), inputHash, nil, 2000000000)
```

## Layer 2 — Pure recompute

Two deterministic computations reproducible off-chain from public inputs:

- **`ComputeTaskHash(stage uint8, taskSeq uint64, inputHash common.Hash, timestamp uint64, expiresAt uint64, prevReplyHashesPacked []byte, workflowRunId common.Hash) (common.Hash, error)`** — the ERC-8301 task hash; the mixed-type `abi.encode` pattern (`uint8, uint256, bytes32, uint256, uint256, bytes32, bytes32`, all static → bare 224-byte concatenation). Pass nil for an empty prev-reply list. Golden vector (recompute-kit `8301/task-hash`, cross-verified against the TypeScript and Rust SDKs): `stage: 1, taskSeq: 0, inputHash: 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8, timestamp: 1700000000, expiresAt: 1700001000, prevReplyHashesPacked: empty, workflowRunId: 0x0000…deadbeef` → `0xf1f404c844a4aff1d0d7d17cebb518a2d386197aad09ab86517eaa01448301ec`.
- **`ComputeReplyHash(outputHash common.Hash, timestamp uint64, replier common.Address, prevTaskHashesPacked []byte, workflowRunId common.Hash) (common.Hash, error)`** — the ERC-8301 reply hash (`bytes32, uint256, address, bytes32, bytes32`; the address left-pads to 32 bytes like any static value). No recompute-kit conformance vector exists for replies yet — the Go golden was computed with the TypeScript SDK's `computeReplyHash` and hard-coded: `outputHash: 0xabcd0000…0000, timestamp: 1700000000, replier: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8, prevTaskHashesPacked: empty, workflowRunId: 0x0000…deadbeef` → `0x65aa5c380a15de82c67b2fb95dacfb12cb327939fd3867c3eb1b64729f17766d`.
- **`ComputeInnerHash(prevHashesPacked []byte) common.Hash`** — the shared inner term: `keccak256(abi.encodePacked(hashesPacked))`; empty → `keccak256("")`, never the zero hash.

The recompute functions are pure `abi.Arguments.Pack` + keccak256 with inline
golden vectors — no RPC, no anvil, no deployed contract required. Errors are
returned, never panicked.

See `client.go` for the contract wrapper, `recompute.go` for the pure
functions, and `go/test/erc8301_integration_test.go` for the testkit
integration test.
