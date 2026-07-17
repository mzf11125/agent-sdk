# agent-sdk

Off-chain SDKs (TypeScript, Python) for ERCs defined in [trustless-ai/agent-ercs](https://github.com/trustless-ai/agent-ercs). Every file in this repo — code, docs, commit messages — is written in English.

## What this repo is

`agent-ercs` defines the on-chain side (Solidity interfaces) of the trustless-ai stack. `agent-sdk` is the off-chain side: client libraries that call those contracts, and — where an ERC's guarantees allow it — independently re-verify a claim off-chain instead of trusting a third party ("recompute-to-verify").

## Adding or updating ERC support

Do not hand-write client code for a new or changed ERC directly. Use the skills:

- **`/add-erc`** — generate TypeScript and Python clients for an ERC that doesn't have SDK support yet. See `.claude/skills/add-erc/SKILL.md`.
- **`/update-erc`** — refresh existing clients after `agent-ercs` changes an ERC's interface. See `.claude/skills/update-erc/SKILL.md`.

Both skills read the interface and README from the `agent-ercs` submodule, judge whether the ERC's claims are recompute-to-verify, and write client code, tests (run against a local `anvil` node via `testkit/`), and a README documenting the verdict.

## Repo layout

```
agent-ercs/                                  # git submodule (tracks `main` by default)
testkit/                                     # shared Foundry harness: deploys contracts to a local anvil node for tests
typescript/src/<category>/<ERCXXXX>/         # TS client + README for each ERC
python/src/agent_sdk/<category>/<ercxxxx>/   # Python client + README for each ERC
```

Directory names mirror `agent-ercs`'s `<category>/<ERCXXXX>` structure. TypeScript keeps the exact casing; Python lowercases the ERC segment per PEP 8.

## Recompute-to-verify

For each ERC, the ERC's own `README.md` (under both the `typescript/` and `python/` implementation directories) records whether its claims can be independently recomputed off-chain from public data, and why. This verdict decides whether the client exposes a `verify()` method and what the tests must cover — see those per-ERC READMEs for the specifics of any given ERC.

## Design background

See `docs/superpowers/specs/2026-07-07-agent-sdk-bootstrap-design.md` for the full design rationale.
