---
name: update-erc
description: Refresh existing TypeScript and Python SDK clients after an ERC's interface or semantics changed in agent-ercs, re-classifying recompute-to-verify capability only if the change affects it.
---

# Update ERC

Refresh SDK support for an ERC that `agent-sdk` already implements, after `agent-ercs` changed something about it.

## Process

1. **Determine what changed.** If not specified, ask which ERC changed in `agent-ercs`, and what changed (or point to the commit/PR). Check out the relevant `agent-ercs` ref if it differs from what's currently checked out.

2. **Diff against what's implemented.** Compare the current `agent-ercs` interface against the existing clients under `typescript/src/<category>/<ERCXXXX>/` and `python/src/agent_sdk/<category>/<ercxxxx>/`.

3. **Re-classify only if warranted.** Re-run the recompute-to-verify classification (see `add-erc`'s step 3) only if the interface or semantics changed in a way that could affect the verdict. Otherwise, carry the existing verdict forward unchanged.

4. **Update the implementation.** Update the affected client code and tests for both languages, and update `testkit/script/<category>/<ERCXXXX>/Deploy<ERCXXXX>.s.sol` (and the mock in `testkit/contracts/mocks/<category>/<ERCXXXX>/`, if one exists) if the contract's constructor or initialization interface changed.

5. **Amend the READMEs.** Append a dated entry to both existing `README.md` files (TS and Python) describing what changed and why. Amend, don't overwrite — the change history is part of the record.

6. **Run every affected test to green** (`forge test`, `npm test`, `pytest`) — run each language's full suite, not just the changed ERC's tests, since shared test infrastructure (one anvil instance and deployer account across all ERCs) can surface cross-ERC issues only when everything runs together.

## What gets committed

Only the updated READMEs, client code, and tests. Discussion during steps 1–3 is scratch and is not committed.
