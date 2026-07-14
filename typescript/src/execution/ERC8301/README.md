# ERC-8301 — Agent Workflow / AgentTask (TypeScript)

**Recompute-to-verify: SPLIT.**

1. **taskHash / replyHash computation — YES.** Both hashes are pure deterministic functions of the struct fields (`abi.encode` of the field values, with an inner `keccak256(abi.encodePacked(...))` for the packed array). The contract exposes the full struct via `getAgentTask()` / `getAgentReply()`, and anyone can independently recompute the hash from those fields without trusting the party who submitted the task or reply.

2. **Workflow FSM correctness — NO.** The FSM transition logic (gate validation, stage progression) is implementation-defined with no interface-level specification. There is no generic recompute check for whether a workflow transition was valid.

## API — Layer 1 (Contract Wrapper)

`AgentWorkflowClient` — wraps `IAgentWorkflow`:

- `run(inputHash, input, expiresAt)` — starts a workflow run; returns `{workflowRunId, taskHash, stage}`
- `result(workflowRunId)` — queries the run's terminal state
- `getTask(taskHash)` — returns `{task, proven}`
- `getReply(replyHash)` — returns `{reply, verifier, proven, verificationDigest}`
- `onAgentReply(reply)` — submits a reply to a dispatched task
- `onAgentProve(replyHashes, proof)` — submits a cryptographic proof for anchored replies

Tests deploy `testkit`'s `MockAgentWorkflow` to a local `anvil` node and call through this client.

## Layer 2 — Pure Recompute

**Pure recompute: YES (two functions).**

Both hash computations are defined in ERC-8301 and can be reproduced off-chain from public inputs as pure stateless functions:

1. **`computeTaskHash(stage, taskSeq, inputHash, timestamp, expiresAt, prevReplyHashesPacked, workflowRunId)`** — computes `taskHash = keccak256(abi.encode(stage, taskSeq, inputHash, timestamp, expiresAt, innerHash, workflowRunId))` where `innerHash = keccak256(abi.encodePacked(prevReplyHashesPacked))`.

   **Critical edge case:** When `prevReplyHashesPacked` is empty (`"0x"`), `innerHash = keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470`. An implementation that special-cases empty to `bytes32(0)` FAILS.

   Tested against golden conformance vector from `recompute-kit/conformance/agent-flow.vectors.json` (step `8301/task-hash`):
   - Inputs: `stage: 1, taskSeq: 0, inputHash: "0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8", timestamp: 1700000000, expiresAt: 1700001000, prevReplyHashesPacked: "0x", workflowRunId: "0x00000000000000000000000000000000000000000000000000000000deadbeef"`
   - Expected: `"0xf1f404c844a4aff1d0d7d17cebb518a2d386197aad09ab86517eaa01448301ec"`

   ABI encode types: `[uint8, uint256, bytes32, uint256, uint256, bytes32, bytes32]`

2. **`computeReplyHash(outputHash, timestamp, replier, prevTaskHashesPacked, workflowRunId)`** — computes `replyHash = keccak256(abi.encode(outputHash, timestamp, replier, innerHash, workflowRunId))` with the same inner hash pattern.

   ABI encode types: `[bytes32, uint256, address, bytes32, bytes32]`

The recompute tests are pure function calls with no RPC, no anvil, and no deployed contract.
