# ERC-8312 — Bounded Agent Actions (Python)

**Recompute-to-verify: YES (two functions).**

## Layer 2 — Pure recompute

Two deterministic computations can be reproduced off-chain from public inputs:

1. **`check_stateful_bound(reserved, confirmed, cap)`** — `(reserved + confirmed) <= cap`. Golden vector: `reserved: 100, confirmed: 0, cap: 150` → `True`.

2. **`check_cursor_headroom(aggregate, cap)`** — `aggregate <= cap`. Golden vector: `aggregate: 0, cap: 8000` → `True`.

These recompute functions are tested against golden conformance vectors from `recompute-kit/conformance/agent-flow.vectors.json` (step `8312/cap-conservation`). The recompute tests are pure function calls with no RPC, no anvil, and no deployed contract.

## Layer 1 — Contract wrappers

Three typed clients for the ERC-8312 interface family:

- **`BoundedAgentActionClient`** — envelope lifecycle: register, advance cursor, set status, read envelope metadata and cursor state.
- **`BudgetSubstrateClient`** — budget profile view functions: `bound`, `spent`, `remaining` (read-only).
- **`ContestableEnvelopeClient`** — contestation: `contest` and `resolve` an envelope.

### Constructor

```python
from agent_sdk.metering.erc8312.client import BoundedAgentActionClient
from eth_account import Account

account = Account.from_key("0x...")
client = BoundedAgentActionClient("http://127.0.0.1:8545", "0xDeployedAddress", account)
```

`BudgetSubstrateClient` does not require an account (read-only).

## ERC-8312 vs. this SDK

| Capability | Interface | SDK Client |
|---|---|---|
| Envelope lifecycle | IBoundedAgentAction | `BoundedAgentActionClient` |
| Budget profile | IBudgetSubstrate | `BudgetSubstrateClient` |
| Contestation | IContestableEnvelope | `ContestableEnvelopeClient` |
| Cap conservation | recompute | `check_stateful_bound`, `check_cursor_headroom` |

## 2026-07-18 — Update from settlement/ to metering/

ERC-8312 was migrated from the `settlement/` category to `metering/` to reflect its new home in `agent-ercs`. Three contract interfaces were added (IBoundedAgentAction, IBudgetSubstrate, IContestableEnvelope) under `contracts/metering/ERC8312/`. The existing recompute functions (`check_stateful_bound`, `check_cursor_headroom`) were moved unchanged from `settlement` to `metering`. All three typed clients were generated fresh.
