# ERC-8312 — Cap Conservation (Python)

**Recompute-to-verify: YES (two functions).**

## Layer 2 — Pure recompute

Two deterministic computations can be reproduced off-chain from public inputs:

1. **`check_stateful_bound(reserved, confirmed, cap)`** — `(reserved + confirmed) <= cap` (ERC-8312 StatefulBound variant). Golden vector: `reserved: 100, confirmed: 0, cap: 150` → `True`.

2. **`check_cursor_headroom(aggregate, cap)`** — `aggregate <= cap` (ERC-8312 Orbmis/headroom variant). Golden vector: `aggregate: 0, cap: 8000` → `True`.

These recompute functions are tested against golden conformance vectors from `recompute-kit/conformance/agent-flow.vectors.json` (step `8312/cap-conservation`). The recompute tests are pure function calls with no RPC, no anvil, and no deployed contract.

## Layer 1 — Contract wrappers

Not yet available — no Solidity interface for ERC-8312 exists in `agent-ercs`. Once an interface is added, run `/update-erc` to generate the contract client.

See `recompute.py` for the implementation and `tests/settlement/erc8312/test_recompute.py` for tests.
