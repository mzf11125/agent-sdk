# agent-sdk Bootstrap Design

**Date**: 2026-07-07
**Status**: Approved (pending spec review)

## 1. Background

[trustless-ai](https://github.com/trustless-ai) builds decentralized AI infrastructure on the principle that every claim must be **trustless**: a third party must be able to independently recompute and verify an outcome from public data, without relying on any trusted intermediary.

[`agent-ercs`](https://github.com/trustless-ai/agent-ercs) is the Solidity side of this stack: interfaces and base implementations for Ethereum ERCs that define agent identity, execution, and verification standards (e.g. ERC-8004 identity, ERC-8274/8299 inference-proof verification, ERC-8301 execution). It organizes contracts as `contracts/<category>/<ERCXXXX>/I*.sol`, with tests mirrored under `test/<category>/<ERCXXXX>/`.

`agent-sdk` (this repo) is the off-chain counterpart: client libraries (TypeScript and Python initially, more languages later) that let a caller interact with contracts implementing these ERCs, and — where the ERC's semantics allow it — independently re-verify a claim off-chain instead of trusting the on-chain caller or a third party.

Each ERC's Solidity interface is a fixed, known shape, so an SDK implementing it is close to mechanical to generate with AI assistance. The one part that isn't mechanical, and is the actual point of this repo, is judging **whether a given ERC's claims are recompute-to-verify** — i.e. whether a caller can independently re-derive and check a result from public inputs, or whether the ERC's guarantee terminates at trusting some signer/oracle with no on-chain recomputation path. That judgment shapes both the generated client code and how it gets tested.

## 2. Goals

- Bootstrap the `agent-sdk` repo structure, tooling, and two Claude Code skills (`add-erc`, `update-erc`) that let a developer generate or refresh a language SDK for any ERC defined in `agent-ercs`.
- Establish one repeatable convention for: repo layout, how the recompute-to-verify judgment is made and recorded, how contracts get deployed for testing, and what "done" looks like for one ERC's SDK.
- Prove the whole pipeline end-to-end by running `add-erc` once, for ERC-8004, producing working TypeScript and Python clients with passing tests.

## 3. Non-goals

- Generating SDKs for every ERC currently in `agent-ercs` (only ERC-8004, as the end-to-end proof).
- A deterministic code-generation engine or templating DSL. The `add-erc`/`update-erc` skills do the generation directly, guided by written conventions; only the mechanical contract-deployment step is factored into a shared helper (see §7).
- Publishing packages to npm/PyPI, or setting up a release pipeline.
- A global cross-ERC registry/dashboard of verify status — each ERC's own README carries that information (see §6).
- Support for languages beyond TypeScript and Python (structure should not preclude adding more later, but none are built now).

## 4. Repo structure

```
agent-sdk/
├── CLAUDE.md                          # onboarding: what this repo is, how to run /add-erc and /update-erc
├── .claude/
│   └── skills/
│       ├── add-erc/SKILL.md
│       └── update-erc/SKILL.md
├── agent-ercs/                        # git submodule, defaults to tracking `main`
├── testkit/                           # shared anvil/forge deployment harness
│   ├── foundry.toml                   # remaps imports to ../agent-ercs
│   └── script/<category>/<ERCXXXX>/Deploy.s.sol
├── typescript/
│   ├── package.json                   # single package, subpath exports per ERC
│   ├── src/<category>/<ERCXXXX>/
│   │   ├── client.ts
│   │   ├── types.ts
│   │   └── README.md                  # recompute-to-verify verdict + rationale for this ERC (TS side)
│   └── test/<category>/<ERCXXXX>/erc.test.ts
└── python/
    ├── pyproject.toml                 # single package: agent_sdk.<category>.<ercxxxx>
    ├── src/agent_sdk/<category>/<ercxxxx>/
    │   ├── client.py
    │   └── README.md                  # recompute-to-verify verdict + rationale for this ERC (Python side)
    └── tests/<category>/<ercxxxx>/test_erc.py
```

Notes:
- Directory names mirror `agent-ercs`'s `<category>/<ERCXXXX>` structure, so any ERC's implementation is easy to locate by analogy. TypeScript keeps `agent-ercs`'s exact casing (`identity/ERC8004`); Python lowercases the ERC segment (`identity/erc8004`) per PEP 8 module-naming convention — same structure, language-idiomatic casing.
- TypeScript and Python are each a single installable package (not one package per ERC), with per-ERC code organized as subpath modules/submodules internally.
- Each ERC gets **two** README files — one under the TS implementation, one under the Python implementation — because the two languages may phrase the verification recomputation differently even though both must agree on the underlying verdict. This is a deliberate duplication, not an oversight.
- No global ERC registry file. Per-ERC READMEs are the only place verify status is recorded (see §6), keeping this consistent with "don't build infrastructure the current scope doesn't need."

## 5. `agent-ercs` dependency

`agent-ercs` is vendored as a **git submodule**, tracking `main` by default. Both skills accept an optional override at invocation time — a different branch, tag, commit, or a local filesystem path — for working against an ERC that hasn't merged to `main` yet. No published-artifact pipeline (npm/PyPI ABI package) is built for this; that would be premature infrastructure for a repo that doesn't have a publishing story yet.

## 6. Recompute-to-verify classification

For every ERC, the `add-erc` skill must produce and record a classification: can a caller independently recompute and check the ERC's central claim off-chain from public data, or does the guarantee terminate at trusting a signer/oracle with no recomputation path?

This affects two things:
1. **Client shape** — if verifiable, the client exposes a `verify(...)` method that performs the off-chain recomputation and returns pass/fail (or throws), independent of trusting the contract's own state. If not verifiable, no such method is generated; the client only exposes plain contract-calling methods.
2. **Test strategy** — if verifiable, tests must include a negative case: construct tampered/incorrect data and assert `verify(...)` rejects it. If not verifiable, tests only need to confirm calls succeed and return expected values.

The verdict and its rationale are recorded in each ERC's `README.md` (both the TS-side and Python-side copies) — not in a separate global file. This README is the durable, committed artifact from the `add-erc`/`update-erc` process; everything else discussed during the skill's question-and-answer flow is scratch and is not committed.

## 7. Shared testkit (the one deterministic piece)

`testkit/` is a small Foundry project whose only job is deploying `agent-ercs` contracts to a local `anvil` node for tests to run against. It remaps Solidity imports to the `agent-ercs` submodule and holds one `Deploy.s.sol` script per ERC (mirroring the same `<category>/<ERCXXXX>` layout).

Both language test suites use the same two mechanical steps, via a small shared helper invoked from each language's test setup:
1. Start a local `anvil` instance.
2. Run `forge script testkit/script/<category>/<ERCXXXX>/Deploy.s.sol --broadcast --rpc-url <anvil-rpc>`, parse the broadcast JSON for the deployed contract address(es).
3. Hand the RPC URL + deployed address(es) to the SDK client under test.

This is the only part of the pipeline that is written once as reusable, deterministic code rather than regenerated per ERC by the skill — because it is identical boilerplate every time, and worth getting right once rather than risking subtle drift between ERCs or between languages.

## 8. `add-erc` skill flow

1. Determine which ERC to add. Ask if not specified. Determine which `agent-ercs` ref to read from (branch/tag/commit/local path); default to `main` if not specified.
2. Read the ERC's interface(s) and `README.md` from `agent-ercs`.
3. Classify recompute-to-verify capability per §6, with explicit reasoning tied to the spec.
4. Propose the client API shape — methods, types, and whether `verify()` is warranted — and get user approval before writing code.
5. Write `README.md` (verdict + rationale + API summary) under both `typescript/src/<category>/<ERCXXXX>/` and `python/src/agent_sdk/<category>/<ercxxxx>/`.
6. Implement the TypeScript and Python clients, add `testkit/script/<category>/<ERCXXXX>/Deploy.s.sol`, and write tests for both languages (deploy via testkit; call client methods; if verifiable, also test that `verify()` rejects tampered data).
7. Run all new tests to green before considering the ERC done.

Mid-flow discussion/scratch content is not committed. Only the final per-ERC `README.md` files, client code, and tests are committed.

## 9. `update-erc` skill flow

Same overall shape as `add-erc`, but diff-driven:
1. Determine which ERC changed in `agent-ercs` and what changed. Ask if not specified.
2. Diff the current `agent-ercs` interface against what's already implemented in `typescript/` and `python/`.
3. Re-run the recompute-to-verify classification **only if** the interface or semantics actually changed in a way that could affect the verdict; otherwise carry the existing verdict forward.
4. Update the affected client code and tests for both languages, and update `testkit/script/<category>/<ERCXXXX>/Deploy.s.sol` if the contract's constructor or initialization interface changed.
5. Append a dated entry to the ERC's existing `README.md` files describing what changed and why — the README is amended with a change log, not silently overwritten.
6. Run tests to green.

## 10. Language conventions

- TypeScript: [`viem`](https://viem.sh) for contract bindings — typesafe, no separate ABI-to-types codegen step needed. Test runner: `vitest`.
- Python: [`web3.py`](https://web3py.readthedocs.io) for contract bindings. Test runner: `pytest`.

These are defaults for the first ERC (ERC-8004); nothing in the repo structure locks in these specific libraries if a future ERC needs something different, but there's no reason to deviate without cause.

## 11. `CLAUDE.md` content

The root `CLAUDE.md` explains, for a developer opening this repo cold:
- What `agent-sdk` is and how it relates to `agent-ercs`.
- That new/changed ERC support is added via the `/add-erc` and `/update-erc` skills, not by hand-writing client code directly.
- The repo layout convention (§4) and where the recompute-to-verify verdict for any given ERC can be found (its `README.md`).
- That everything in this repo, including code, docs, and commit messages, is written in English.

## 12. First deliverable / acceptance criteria

1. The repo skeleton in §4 exists: `CLAUDE.md`, `.claude/skills/{add-erc,update-erc}/SKILL.md`, the `agent-ercs` submodule, `testkit/` with its `foundry.toml`, and the empty `typescript/` and `python/` package scaffolding (package manifests, no ERC content yet).
2. Running `add-erc` once for ERC-8004 produces, and commits:
   - `typescript/src/identity/ERC8004/{client.ts,types.ts,README.md}` and a passing `test/identity/ERC8004/erc.test.ts`.
   - `python/src/agent_sdk/identity/erc8004/{client.py,README.md}` and a passing `tests/identity/erc8004/test_erc.py`.
   - `testkit/script/identity/ERC8004/Deploy.s.sol`.
3. Both test suites pass by deploying ERC-8004's contracts to a local `anvil` node via the shared testkit and calling through the generated clients.
