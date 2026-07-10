---
name: add-erc
description: Generate TypeScript and Python SDK clients for an ERC defined in agent-ercs that doesn't have SDK support yet, including a recompute-to-verify classification and tests run against a local anvil deployment.
---

# Add ERC

Generate off-chain client support for one ERC that `agent-sdk` doesn't yet implement.

## Process

1. **Determine which ERC.** If not specified, ask which ERC (by number, e.g. "ERC-8301") to add. If the `agent-ercs` submodule should be read from something other than its currently checked-out `main`, ask for the branch, tag, commit, or local path to use, and check that out in `agent-ercs/` before continuing (default: whatever `main` currently points to).

2. **Read the spec.** Read the ERC's interface file(s) and `README.md` under `agent-ercs/contracts/<category>/<ERCXXXX>/`.

3. **Classify recompute-to-verify capability.** An ERC can make more than one distinct claim — classify each central claim separately rather than forcing one verdict for the whole ERC. For each claim, determine: can a caller independently obtain the same authoritative answer without trusting whoever originally submitted it — either by recomputing off-chain from public data, or by calling a deterministic, callable-by-anyone, immutable on-chain check themselves — or does the guarantee terminate at trusting a specific deployment's unfixed convention (a signing scheme, a derived value) with no way to even know what to check? Write out the reasoning — this is the part of the job that isn't mechanical.
   - See `typescript/src/identity/ERC8004/README.md` for a clean NOT-verifiable case: the interface leaves a signing convention completely unfixed, so there's nothing a generic SDK can check at all.
   - See `typescript/src/verify/ERC8274/README.md` for a split verdict on one ERC: the core validity check *is* recompute-to-verify (anyone can call the deployed, immutable verifier contract themselves and get an authoritative answer — that's a real instance of recompute-to-verify, not "just asking the same contract again"), while a separate derived value (an audit-trail digest) is NOT, because one of its inputs isn't exposed anywhere in the interface. Don't let one claim's verdict force the other's.

4. **Propose the client API.** Based on the interface and the classification, propose the method list for both languages — including whether a `verify()` method is warranted — and get the user's approval before writing any code.

5. **Write the per-ERC READMEs.** Under both `typescript/src/<category>/<ERCXXXX>/README.md` and `python/src/agent_sdk/<category>/<ercxxxx>/README.md` (lowercase ERC segment for Python), record the verdict, its rationale, and the API summary.

6. **Implement.** For each language:
   - Hand-write the ABI fragment for the functions/events the client uses, matching the interface exactly (no dynamic codegen from build artifacts).
   - Implement the client, following the shape and conventions of `typescript/src/identity/ERC8004/client.ts` / `python/src/agent_sdk/identity/erc8004/client.py` for a single-contract ERC, or `typescript/src/verify/ERC8274/*Client.ts` / `python/src/agent_sdk/verify/erc8274/client.py` for an ERC that's really several interfaces meant to be deployed as separate, cross-referencing contracts — don't force a multi-contract ERC into one client class. For a claim classified as recompute-to-verify (a deterministic, callable-by-anyone check), expose it as a read-only simulated call/`.call()` rather than a broadcast transaction — nobody should need to spend gas or hold a funded key just to check something.
   - If the ERC needs a contract to deploy for testing and `agent-ercs` has no base implementation yet, write a minimal reference implementation under `testkit/contracts/mocks/<category>/<ERCXXXX>/` (one file per contract if the ERC needs more than one), clearly commented as local-testing-only (see `MockIdentityRegistry.sol` for a single-contract pattern, `MockProofVerifier.sol`/`MockAgentVerifier.sol`/`MockAgentVerifiable.sol` for a multi-contract one), plus a Foundry unit test for it/them under `testkit/test/<category>/<ERCXXXX>/`.
   - Write `testkit/script/<category>/<ERCXXXX>/Deploy<ERCXXXX>.s.sol` (file basename must match its contract name, e.g. `DeployERC8301.s.sol` containing `contract DeployERC8301` — Foundry keys broadcast artifacts by script basename only, so reusing a generic name like `Deploy.s.sol` across ERCs would collide). If the ERC needs several wired-together contracts, deploy all of them in one script (constructor-inject each into the next) — `testkit/scripts/deploy.sh` prints one address per line in the order each was deployed; use `deployContracts()`/`deploy_contracts()` (plural, returning the full list) instead of the single-address `deployContract()`/`deploy_contract()` to receive all of them (see `typescript/test/verify/ERC8274/erc.test.ts` / `python/tests/verify/erc8274/test_erc.py`).
   - Write tests for both languages that deploy via `testkit/scripts/deploy.sh` (see `typescript/test/identity/ERC8004/erc.test.ts` and `python/tests/identity/erc8004/test_erc.py` for the single-contract wiring pattern, or the ERC-8274 test files above for multi-contract) and call the client's methods. For any claim classified as recompute-to-verify, also test that the check rejects tampered/incorrect data (a bad proof, a bad signature) — some checks reject by returning a falsy result rather than reverting; assert whichever the contract actually does, don't assume a revert.
   - If double-checking a byte-encoding assumption against the actual Solidity (e.g. whether a hash was built with `abi.encode` vs `abi.encodePacked` — they differ for `bool` and other non-32-byte-aligned types), verify it against the real contract rather than assuming the two are interchangeable.

7. **Run every new test to green** (`forge test`, `npm test`, `pytest`) before considering the ERC done. If more than one ERC's tests now exist for a language, run that language's *full* suite, not just the new files in isolation — shared test infrastructure (like one anvil instance and deployer account across all ERCs) can only reveal cross-file issues, such as a nonce race from parallel test execution, when everything runs together.

## What gets committed

Only the final READMEs, client code, and tests. Discussion during steps 1–4 is scratch and is not committed.
