# ERC-8301 — Agent Workflow / AgentTask (Python)

**Recompute-to-verify: SPLIT.**

1. **taskHash / replyHash computation — YES.** Both hashes are pure deterministic functions of the struct fields (`abi.encode` of the field values, with an inner `keccak256(abi.encodePacked(...))` for the packed array). The contract exposes the full struct via `getAgentTask()` / `getAgentReply()`, and anyone can independently recompute the hash from those fields without trusting the party who submitted the task or reply.

2. **Workflow FSM correctness — NO.** The FSM transition logic (gate validation, stage progression) is implementation-defined with no interface-level specification. There is no generic recompute check for whether a workflow transition was valid.

## API — Layer 1 (Contract Wrapper)

`AgentWorkflowClient` — wraps `IAgentWorkflow`:

- `run(input_hash, input_data, expires_at)` — starts a workflow run
- `result(workflow_run_id)` — queries the run's terminal state
- `get_task(task_hash)` — returns `TaskResult`
- `get_reply(reply_hash)` — returns `ReplyResult`
- `on_agent_reply(reply)` — submits a reply to a dispatched task
- `on_agent_prove(reply_hashes, proof)` — submits a cryptographic proof for anchored replies

## Layer 2 — Pure Recompute

**Pure recompute: YES (two functions).**

1. **`compute_task_hash(...)`** — see TypeScript README for the identical spec and golden vectors
2. **`compute_reply_hash(...)`** — see TypeScript README for the identical spec

Tests are pure function calls with no RPC, no anvil, and no deployed contract.
