# agent-sdk

Off-chain SDKs (TypeScript, Python) for the ERCs defined in [trustless-ai/agent-ercs](https://github.com/trustless-ai/agent-ercs).

> **This SDK is AI-generated, not hand-written.** Every ERC's client code, tests, and verification-capability writeup in this repo was produced by an AI coding agent (Claude Code) following the `/add-erc` and `/update-erc` skills below — not typed by a human line by line. Adding support for a new ERC is a single prompt, not a pull request someone hand-rolls. See [How the SDKs get written](#how-the-sdks-get-written) for what that actually means and why it's safe to trust.

## What this is

[trustless-ai](https://github.com/trustless-ai) builds decentralized AI infrastructure on one rule: every claim must be independently verifiable from public data, with no trusted middleman. `agent-ercs` is the on-chain half of that stack — Solidity interfaces for agent identity, execution, and verification. `agent-sdk` (this repo) is the off-chain half: client libraries that call those contracts and, where an ERC's guarantees actually allow it, let a caller **recompute and check a claim itself** instead of trusting the chain, the contract deployer, or this SDK.

## Supported ERCs

| ERC | Category | Recompute-to-verify | TypeScript | Python |
|-----|----------|----------------------|------------|--------|
| [ERC-8004](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/identity/ERC8004) — Identity Registry | `identity` | ❌ No ([why](typescript/src/identity/ERC8004/README.md)) | [`typescript/src/identity/ERC8004`](typescript/src/identity/ERC8004) | [`python/src/agent_sdk/identity/erc8004`](python/src/agent_sdk/identity/erc8004) |
| [ERC-8274](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/verify/ERC8274) — AI Inference Proof Verification | `verify` | ⚠️ Split ([why](typescript/src/verify/ERC8274/README.md)) | [`typescript/src/verify/ERC8274`](typescript/src/verify/ERC8274) | [`python/src/agent_sdk/verify/erc8274`](python/src/agent_sdk/verify/erc8274) |

More ERCs get added the same way every ERC above did — see [Adding a new ERC](#adding-a-new-erc).

## Using the SDKs

Neither package is published yet; use them from a local checkout.

**TypeScript** (the package isn't published yet — import it by relative path from within this repo, the same way its own tests do)

```bash
cd typescript
npm install
```

```ts
import { privateKeyToAccount } from 'viem/accounts'
import { IdentityRegistryClient } from './src/identity/ERC8004/client.js'

const account = privateKeyToAccount(myPrivateKey)
const client = new IdentityRegistryClient({ rpcUrl, address: deployedAddress }, account)

const agentId = await client.register('ipfs://my-agent-card')
```

**Python**

```bash
cd python
python3 -m venv .venv && source .venv/bin/activate
pip install -e '.[dev]'
```

```python
from eth_account import Account
from agent_sdk.identity.erc8004.client import IdentityRegistryClient

account = Account.from_key(my_private_key)
client = IdentityRegistryClient(rpc_url, deployed_address, account)

agent_id = client.register("ipfs://my-agent-card")
```

Each ERC's client is documented in its own `README.md` next to the code (linked in the table above) — that's where you'll find the full method list and, importantly, whether `verify()` exists for that ERC and why.

## Running the tests

Every ERC's tests deploy a local, testing-only reference contract to `anvil` (Foundry's local node) and call through the real generated client — nothing is mocked.

```bash
# Solidity (the reference contracts used for testing)
cd testkit && forge test

# TypeScript
cd typescript && npm test

# Python
cd python && source .venv/bin/activate && pytest
```

Requires `forge`/`cast`/`anvil` ([Foundry](https://book.getfoundry.sh/getting-started/installation)), `jq`, `curl`, Node.js, and Python 3.10+.

## How the SDKs get written

`agent-ercs` interfaces are stable, fully-specified Solidity — mechanically translating one into a TypeScript/Python client is exactly the kind of task an AI coding agent does reliably. The part that isn't mechanical is judgment: **can a caller independently recompute this ERC's claim from public data, or does it terminate at trusting a signer with no on-chain recomputation path?** That classification decides whether the generated client gets a `verify()` method and what the tests have to cover — and it's reasoned through and written down, not assumed.

This is why the SDK is AI-generated rather than hand-written: the mechanical part (bindings, boilerplate, wiring tests to a local chain) is handled consistently every time by following the same skill, and the judgment part is done explicitly, with its reasoning committed to the repo (see each ERC's `README.md`) instead of living only in a PR description that rots.

## Adding a new ERC

Open this repo in Claude Code and run:

```
/add-erc
```

Tell it which ERC (e.g. "ERC-8301"), or let it ask. It will:

1. Read the ERC's interface and README from the `agent-ercs` submodule.
2. Judge — with stated reasoning — whether the ERC's claims are recompute-to-verify.
3. Propose the client's method list for your approval before writing anything.
4. Write the TypeScript and Python clients, a local test contract if one doesn't exist yet in `agent-ercs`, and tests that deploy and call through the real client on a local `anvil` node.
5. Run every new test to green.

If `agent-ercs` changes an ERC you already support, run `/update-erc` instead — it diffs against what's implemented and only re-does the classification if the change actually affects it.

Both skills are defined in [`.claude/skills/add-erc/SKILL.md`](.claude/skills/add-erc/SKILL.md) and [`.claude/skills/update-erc/SKILL.md`](.claude/skills/update-erc/SKILL.md).

## Repo layout

```
agent-ercs/                                  # git submodule: the upstream Solidity interfaces (tracks `main`)
testkit/                                     # shared Foundry harness: local test contracts + anvil deployment scripts
typescript/src/<category>/<ERCXXXX>/         # TS client + README for each ERC
python/src/agent_sdk/<category>/<ercxxxx>/   # Python client + README for each ERC
```

Directory names mirror `agent-ercs`'s `<category>/<ERCXXXX>` structure; TypeScript keeps the exact casing, Python lowercases the ERC segment per PEP 8.

## Further reading

- [`CLAUDE.md`](CLAUDE.md) — the same information, written for the AI agent operating this repo rather than a human reader.
- [`docs/superpowers/specs/2026-07-07-agent-sdk-bootstrap-design.md`](docs/superpowers/specs/2026-07-07-agent-sdk-bootstrap-design.md) — the full design rationale for why this repo is structured the way it is.
