<div align="center">
    <img width="3265" height="994" alt="e89e95aa8ed6a99833394682c6632cb7" src="https://github.com/user-attachments/assets/9ed2ab1a-aab5-4133-a5e2-5bcbdad0ffea" />
</div>

# agent-sdk

Off-chain SDKs (TypeScript, Python) for the ERCs defined in [trustless-ai/agent-ercs](https://github.com/trustless-ai/agent-ercs).

## What this is

`agent-sdk` is the off-chain half of the [trustless-ai](https://github.com/trustless-ai) stack — `agent-ercs` defines the on-chain Solidity interfaces, and this repo provides the client libraries that talk to them. It does two things:

- **Call contracts** — typed clients that read and write deployed ERC contracts, wrapping `viem` (TypeScript) / `web3.py` (Python) with the exact types and method names from each ERC interface.
- **Recompute claims** — pure stateless functions that reproduce ERC-defined cryptographic operations (hashing, padding, encoding, arithmetic) from public inputs, without touching a chain. Every recompute function is tested against golden conformance vectors so you can verify a claim locally, offline, or in a browser.

> **This SDK is AI-generated, not hand-written.** Every ERC's client code, tests, and documentation is produced by an AI coding agent following the `/add-erc` skill — not typed by a human line by line. See [Adding a new ERC](#adding-a-new-erc).

---

## Supported ERCs

| ERC | Category | Contract Calls | Recompute |
|-----|----------|----------------|-----------|
| [ERC-8263](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/anchor/ERC8263) — OnChain Proof Anchor | `anchor` | `OnChainProofClient` | — |
| [ERC-8004](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/identity/ERC8004) — Identity Registry | `identity` | `IdentityRegistryClient` | `computeAgentId` |
| [ERC-8323](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/identity/ERC8323) — Source-Token Agent Binding | `identity` | `SourceBindingClient` | — |
| [ERC-8274](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/verify/ERC8274) — AI Inference Proof | `verify` | `ProofVerifierClient`, `AgentVerifierClient`, `getTrustedVerifier` | — |
| [ERC-8281](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/verify/ERC8281) — Observation Commitment Protocol (OCP) | `verify` | `ObservationCommitmentClient` | `computeObservationDigest` |
| [ERC-8299](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/verify/ERC8299) — WYRIWE Input Provenance | `verify` | `WyriweAttestationClient`, `JudgmentExecutionClient` | `computeRawInputHash`, `computeSanitizationPipelineHash` |
| [ERC-8301](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/execution/ERC8301) — Agent Execution | `execution` | `AgentWorkflowClient` | `computeTaskHash`, `computeReplyHash` |
| [ERC-8203](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/settlement/ERC8203) — Consult Escrow | `settlement` | `ConsultEscrowClient` | `computeVerdictHash` |
| [ERC-8275](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/reputation/ERC8275) — Agent Reputation | `reputation` | `AgentReputationClient` | `computeWinRate` |
| [ERC-8312](https://github.com/trustless-ai/agent-ercs/tree/main/contracts/settlement/ERC8312) — Cap Conservation | `settlement` | — | `checkStatefulBound`, `checkCursorHeadroom` |

> **Contract Calls** are typed clients that call deployed contracts. **Recompute** functions are pure, stateless, and run without an RPC endpoint — they reproduce the same deterministic computation the contract performs, verified against golden conformance vectors.

---

## Using the SDKs

### TypeScript

```bash
npm install @trustless-ai/agent-sdk
```

```ts
// Contract call — talk to a deployed ERC contract
import { SomeClient } from '@trustless-ai/agent-sdk/<category>/<ERCXXXX>'

const client = new SomeClient({ rpcUrl, address: deployedAddress }, account)
const result = await client.someMethod(...)

// Recompute — verify a claim without touching the chain
import { someHash } from '@trustless-ai/agent-sdk/<category>/<ERCXXXX>/recompute'

const hash = someHash(publicInput)
```

### Python

```bash
cd python
python3 -m venv .venv && source .venv/bin/activate
pip install -e '.[dev]'
```

```python
# Contract call
from agent_sdk.<category>.<ercxxxx>.client import SomeClient

client = SomeClient(rpc_url, deployed_address, account)
result = client.some_method(...)

# Recompute
from agent_sdk.<category>.<ercxxxx>.recompute import some_hash

hash = some_hash(public_input)
```

Each ERC's own `README.md` documents the full method list, types, and whether `verify()` is available for that ERC.

---

## Adding a new ERC

Open this repo in Claude Code and run:

```
/add-erc
```

Tell it which ERC number, or let it ask. It will:

1. Read the ERC's interface and README from the `agent-ercs` submodule.
2. Judge which capability applies — contract calls, recompute, or both — with stated reasoning.
3. Propose the client API for your approval before writing any code.
4. Write the TypeScript and Python clients, recompute functions (where applicable), test contracts, and tests.
5. Run every test to green.

If `agent-ercs` changes an ERC you already support, run `/update-erc` instead — it diffs against what's implemented and only re-does the classification if the change affects it.

Both skills are defined in [`.claude/skills/add-erc/SKILL.md`](.claude/skills/add-erc/SKILL.md) and [`.claude/skills/update-erc/SKILL.md`](.claude/skills/update-erc/SKILL.md). This is also how you contribute to this repo — the skills are the contribution pipeline.

---

## License

Apache 2.0 — see [LICENSE](./LICENSE).
