<div align="center">
    <img width="3265" height="994" alt="e89e95aa8ed6a99833394682c6632cb7" src="https://github.com/user-attachments/assets/9ed2ab1a-aab5-4133-a5e2-5bcbdad0ffea" />
</div>

# agent-sdk

Off-chain SDKs (TypeScript, Python) for the ERCs defined in [trustless-ai/agent-ercs](https://github.com/trustless-ai/agent-ercs).

[trustless-ai](https://github.com/trustless-ai) builds decentralized AI infrastructure on one rule: every claim must be independently verifiable from public data, with no trusted middleman. `agent-ercs` is the on-chain half of that stack — Solidity interfaces for agent identity, execution, and verification. `agent-sdk` (this repo) is the off-chain half: typed client libraries that call those contracts and, where an ERC's guarantees allow it, pure functions that **recompute** the ERC's cryptographic operations from public inputs so you don't need the chain to check a claim.

> **This SDK is AI-generated, not hand-written.** Every ERC's client code, tests, and capability writeup in this repo was produced by an AI coding agent (Claude Code) following the `/add-erc` and `/update-erc` skills — not typed by a human line by line. Adding support for a new ERC is a single prompt, not a pull request someone hand-rolls. See [How the SDKs get written](#how-the-sdks-get-written) for what that actually means and why it's safe to trust.

---

## Two Capabilities

Every ERC in this SDK exposes one or both of these orthogonal capabilities. Which ones an ERC supports depends on whether its claims can be independently recomputed from public data or terminate at an on-chain signature.

### Contract Calls

Typed clients that call deployed ERC contracts — read state, send transactions, and verify proofs by querying the chain. Every client wraps the underlying `viem` (TypeScript) / `web3.py` (Python) calls with the exact types and method names from the ERC interface.

```ts
import { IdentityRegistryClient } from 'agent-sdk/identity/ERC8004'

const client = new IdentityRegistryClient({ rpcUrl, address }, account)
const agentId = await client.register('ipfs://my-agent-card')
```

### Recompute

Pure stateless functions that reproduce ERC-defined cryptographic operations from public inputs — hashing, padding, arithmetic, encoding. No RPC, no account, no chain needed. Every recompute function is tested against golden conformance vectors embedded in the test suite, so you can:

- Verify a hash locally before deciding to look up an on-chain proof
- Double-check a contract's output without trusting the RPC provider
- Use the logic offline, in a browser, or in a CLI tool without ever touching the chain

```ts
import { computeAgentId } from 'agent-sdk/identity/ERC8004/recompute'

const id = computeAgentId(42)
// → "0x000000000000000000000000000000000000000000000000000000000000002a"
```

---

## Supported ERCs

| ERC | Category | Contract Calls | Recompute | Recompute-to-verify |
|-----|----------|:---:|:---:|----------------------|
| [ERC-8004](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/identity/ERC8004) — Identity Registry | `identity` | `IdentityRegistryClient` | `computeAgentId` (zero-pad) | ❌ No |
| [ERC-8274](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/verify/ERC8274) — AI Inference Proof Verification | `verify` | `ProofVerifierClient`, `AgentVerifierClient`, `getTrustedVerifier` | — (contract-call verification) | ⚠️ Split |
| [ERC-8299](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/verify/ERC8299) — WYRIWE Input Provenance | `verify` | `WyriweAttestationClient`, `JudgmentExecutionClient` | `computeRawInputHash`, `computeSanitizationPipelineHash` | ✅ Yes |
| [ERC-8301](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/execution/ERC8301) — Agent Execution | `execution` | `AgentWorkflowClient` | `computeTaskHash`, `computeReplyHash` | ⚠️ Split |
| [ERC-8203](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/settlement/ERC8203) — Consult Escrow | `settlement` | `ConsultEscrowClient` | `computeVerdictHash` | ✅ Yes |
| [ERC-8275](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/settlement/ERC8275) — Win Rate | `settlement` | — | `computeWinRate` | — |
| [ERC-8312](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/settlement/ERC8312) — Cap Conservation | `settlement` | — | `checkStatefulBound`, `checkCursorHeadroom` | — |

Each ERC's own README documents whether `verify()` exists for that ERC, whether it's recompute-to-verify, and the full method list.

---

## Architecture

```
agent-ercs (Solidity interfaces)
     │
     ▼
agent-sdk (this repo — TypeScript + Python clients + recompute functions)
     │
     ├──────────────────────┬──────────────────────┐
     ▼                      ▼                      ▼
recompute-lens       recompute-kit         onchain-boiler-kit
(browser verify)     (CLI conformance)     (full-stack dapp)
```

- **recompute-lens** — a browser-based tool that runs recompute functions client-side to verify on-chain claims without an RPC endpoint. Pure JS, no node needed.
- **recompute-kit** — a CLI toolkit that drives recompute functions against golden conformance vectors. The same vectors used in this SDK's CI tests guarantee that every recompute function is correct by construction.
- **onchain-boiler-kit** — full-stack dapp templates that wire the SDK's contract clients into a UI, wallet connection, and deployment pipeline.

---

## The Recompute Guarantee

> Every recompute function in this SDK reproduces the exact output of its corresponding ERC specification. Golden conformance vectors from recompute-kit are embedded in every recompute test — if the vector matches, the function is correct by construction. If it doesn't, the test fails at CI.

This guarantee is why a recompute function can be used independently of the chain: it's not an approximation, not a simulation — it is the same deterministic computation the contract would perform, re-derived from public inputs.

### Quick example

A settlement verdict in ERC-8203 is stored on-chain as a commitment hash. Instead of trusting an RPC to return the right hash, you recompute it yourself:

```ts
import { computeVerdictHash } from 'agent-sdk/settlement/ERC8203/recompute'

const hash = computeVerdictHash(
  '0xabc...',           // jobId from the on-chain Released event
  'task completed OK',  // resultText from the same event
)
// Compare with the commitment hash emitted in the event
```

If `hash` matches what the contract emitted, the verdict is authentic — no RPC provider, no third party, just the public data and this SDK.

---

## Using the SDKs

Neither package is published yet; use them from a local checkout.

### TypeScript

```bash
cd typescript
npm install
```

**Contract call:**
```ts
import { privateKeyToAccount } from 'viem/accounts'
import { IdentityRegistryClient } from 'agent-sdk/identity/ERC8004'

const account = privateKeyToAccount(myPrivateKey)
const client = new IdentityRegistryClient({ rpcUrl, address: deployedAddress }, account)
const agentId = await client.register('ipfs://my-agent-card')
```

**Recompute:**
```ts
import { computeAgentId } from 'agent-sdk/identity/ERC8004/recompute'
const id = computeAgentId(registryId)
```

### Python

```bash
cd python
python3 -m venv .venv && source .venv/bin/activate
pip install -e '.[dev]'
```

**Contract call:**
```python
from eth_account import Account
from agent_sdk.identity.erc8004.client import IdentityRegistryClient

account = Account.from_key(my_private_key)
client = IdentityRegistryClient(rpc_url, deployed_address, account)
agent_id = client.register("ipfs://my-agent-card")
```

**Recompute:**
```python
from agent_sdk.identity.erc8004.recompute import compute_agent_id
id = compute_agent_id(registry_id)
```

---

## Running the tests

Every ERC's tests deploy a local, testing-only reference contract to `anvil` (Foundry's local node) and call through the real generated client — nothing is mocked. Recompute tests run against embedded golden conformance vectors.

```bash
# Solidity (the reference contracts used for testing)
cd testkit && forge test

# TypeScript
cd typescript && npm test

# Python
cd python && source .venv/bin/activate && pytest
```

Requires `forge`/`cast`/`anvil` ([Foundry](https://book.getfoundry.sh/getting-started/installation)), `jq`, `curl`, Node.js, and Python 3.10+.

---

## How the SDKs get written

`agent-ercs` interfaces are stable, fully-specified Solidity — mechanically translating one into a TypeScript/Python client is exactly the kind of task an AI coding agent does reliably. The part that isn't mechanical is judgment: **can a caller independently recompute this ERC's claim from public data, or does it terminate at trusting a signer with no on-chain recomputation path?** That classification decides whether the generated client gets a `verify()` method, whether a recompute module is generated, and what the tests have to cover — and it's reasoned through and written down, not assumed.

This is why the SDK is AI-generated rather than hand-written: the mechanical part (bindings, boilerplate, wiring tests to a local chain) is handled consistently every time by following the same skill, and the judgment part is done explicitly, with its reasoning committed to the repo (see each ERC's `README.md`) instead of living only in a PR description that rots.

---

## Adding a new ERC

Open this repo in Claude Code and run:

```
/add-erc
```

Tell it which ERC (e.g. "ERC-8301"), or let it ask. It will:

1. Read the ERC's interface and README from the `agent-ercs` submodule.
2. Judge — with stated reasoning — whether the ERC's claims are recompute-to-verify and which capability (contract calls, recompute, or both) applies.
3. Propose the client's method list for your approval before writing anything.
4. Write the TypeScript and Python clients (contract calls + recompute functions where applicable), a local test contract if one doesn't exist yet in `agent-ercs`, and tests that either deploy and call through the real client on a local `anvil` node or run recompute functions against golden conformance vectors.
5. Run every new test to green.

If `agent-ercs` changes an ERC you already support, run `/update-erc` instead — it diffs against what's implemented and only re-does the classification if the change actually affects it.

Both skills are defined in [`.claude/skills/add-erc/SKILL.md`](.claude/skills/add-erc/SKILL.md) and [`.claude/skills/update-erc/SKILL.md`](.claude/skills/update-erc/SKILL.md).

---

## Repo layout

```
agent-ercs/                                  # git submodule: the upstream Solidity interfaces (tracks `main`)
testkit/                                     # shared Foundry harness: local test contracts + anvil deployment scripts
typescript/src/<category>/<ERCXXXX>/         # TS client + README for each ERC
typescript/src/<category>/<ERCXXXX>/recompute.ts   # TS recompute functions (when applicable)
python/src/agent_sdk/<category>/<ercxxxx>/         # Python client + README for each ERC
python/src/agent_sdk/<category>/<ercxxxx>/recompute.py   # Python recompute functions (when applicable)
```

Directory names mirror `agent-ercs`'s `<category>/<ERCXXXX>` structure; TypeScript keeps the exact casing, Python lowercases the ERC segment per PEP 8.

---

## Contributing

### Submit an ERC

If you have an ERC that fits the trustless AI stack, start by contributing to [`agent-ercs`](https://github.com/trustless-ai/agent-ercs) — the interface comes first. Once it's merged, the `/add-erc` skill in this repo generates the off-chain SDK automatically.

### Report a bug or improve something

Found an issue in a client, a spec mismatch, or a gap in the docs? [Open an issue](https://github.com/trustless-ai/agent-sdk/issues) or send a PR. Bug reports, test contributions, and documentation improvements all count.

### Join the conversation

Hop into the [Telegram group](https://t.me/+rKbR1EQcT8QxNzI0) to brainstorm, ask questions, or just see what people are working on.

---

## Further reading

- [`CLAUDE.md`](CLAUDE.md) — the same information, written for the AI agent operating this repo rather than a human reader.
- [`docs/superpowers/specs/2026-07-07-agent-sdk-bootstrap-design.md`](docs/superpowers/specs/2026-07-07-agent-sdk-bootstrap-design.md) — the full design rationale for why this repo is structured the way it is.
- [agent-ercs](https://github.com/trustless-ai/agent-ercs) — the on-chain Solidity interfaces this SDK implements.
