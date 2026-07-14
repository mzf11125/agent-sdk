# ERC-8275 — Reputation (Python)

**Recompute-to-verify: YES (one function).**

## Layer 2 — Pure recompute

One deterministic computation can be reproduced off-chain from public inputs:

- **`compute_win_rate(wins, losses)`** — `winRate = gated_wins / (gated_wins + gated_losses)`, rounded to 4 decimal places (ERC-8275). Golden vector: `wins: 16, losses: 15` → `0.5161`.

The recompute function is tested against golden conformance vectors from `recompute-kit/conformance/agent-flow.vectors.json` (step `8275/reputation`). The recompute tests are pure function calls with no RPC, no anvil, and no deployed contract.

## Layer 1 — Contract wrappers

Not yet available — no Solidity interface for ERC-8275 exists in `agent-ercs`. Once an interface is added, run `/update-erc` to generate the contract client.

See `recompute.py` for the implementation and `tests/settlement/erc8275/test_recompute.py` for tests.
