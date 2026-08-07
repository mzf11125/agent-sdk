# Integrations

Third-party tools that compose with the trustless-ai ERC stack.

**These live in their maintainers' own repositories.** This file is an index, not an endorsement:
listing a project means it targets this stack and its maintainer keeps it current — nothing more.
Anything hosted here would inherit this SDK's implicit warranty, and neither we nor the vendor
benefit from that confusion.

## How to add yours

Open a PR adding one row. Requirements:

1. **Your integration lives in your repo**, so the people with the incentive to keep it working own it.
2. **One line describing what it adds**, in plain terms — capabilities, not marketing.
3. **A composition boundary that respects the stack's**: an integration may sit *beside* the ERC
   primitives, never *above* them. The verifier decides validity; identity, anchoring and proof
   semantics stay where the ERCs put them.
4. **Say what a reader can check.** If your integration produces a claim, state how someone
   verifies it independently (a signed artifact, an on-chain record, a published vector). "Our API
   returned OK" is not a check — and an integration whose claims are only assertable is welcome
   here, it just says so plainly.

| project | what it adds | maintainer | repo |
|---|---|---|---|
| Bastion | A policy gate beside the ERC stack. It runs pre-execution simulation, programmable rules, and a human in the loop step, then emits the allow or deny as a zero knowledge verdict (CAPV, draft ERC-8354) that an on chain Guard consumes, so a reader checks the gate ran without trusting the API. | ZKOS Labs | https://github.com/zkos-labs/bastion |

## Building an integration

The SDK's own surface is the contract: ERC clients for reads and writes, and pure recompute
functions for anything derivable. If your tool needs to state something the SDK can already
derive, derive it — don't re-assert it. Cross-lane vectors live in `testkit/vectors/` and in
`recompute-kit/conformance/`; an integration that computes the same values is welcome to check
itself against them.
