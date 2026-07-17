# ERC-8004 — Identity Registry (Python)

**Recompute-to-verify: NO.**

`IIdentityRegistry`'s registration and metadata functions (`register`, `set_agent_uri`, `set_metadata`, `get_metadata`) are plain contract state — the contract itself is the source of truth, there is nothing to independently recompute from public data. `unset_agent_wallet`/`get_agent_wallet` are likewise plain state.

`set_agent_wallet` is signature-gated (EIP-712/ERC-1271), which looks like a recompute-to-verify candidate — a caller could, in principle, independently check that a signature proves control of the agent's wallet without trusting the contract's own check. However, the interface does **not** fix the EIP-712 domain (name/version) or the typed-data struct/typehash used for that signature — those are implementation details left to whoever deploys a concrete `IIdentityRegistry`. A generic SDK cannot recompute a check whose exact shape isn't guaranteed by the ERC itself, so no `verify()` method is generated here. (Same verdict as the TypeScript implementation — see its README for the full rationale.)

If `agent-ercs` publishes a base implementation that fixes this convention, revisit this verdict via `/update-erc`.

## API

- `register(agent_uri="", metadata=None)` → `agent_id`
- `set_agent_uri(agent_id, agent_uri)`
- `get_agent_uri(agent_id)` (reads via the standard ERC-721 `tokenURI`)
- `get_metadata(agent_id, metadata_key)` / `set_metadata(agent_id, metadata_key, metadata_value)`
- `set_agent_wallet(agent_id, new_wallet, deadline, signature)` / `get_agent_wallet(agent_id)` / `unset_agent_wallet(agent_id)`
- `owner_of(agent_id)` (standard ERC-721)

Tests deploy `testkit`'s `MockIdentityRegistry` (a reference implementation for local testing only — see `agent-ercs`'s README on interface vs. base implementation vs. example/reference contracts) to a local `anvil` node and call through this client.

## Layer 2 — Pure recompute (added 2026-07-14)

**Pure recompute: YES (one function).**

Even though ERC-8004 is NOT recompute-to-verify (the signing convention is unfixed), one trivial deterministic computation can be reproduced off-chain:

- `compute_agent_id(registry_id)` — `agentId = bytes32(uint256(registryId))`, a left-padded zero-extension of the registry id (no hash).

This lives in `recompute.py` and is tested against golden conformance vectors from `recompute-kit/conformance/agent-flow.vectors.json` (step `8004/agent-id`). The recompute tests are pure function calls with no RPC, no anvil, and no deployed contract.

See `recompute.py` for the implementation and `tests/identity/erc8004/test_recompute.py` for tests.
