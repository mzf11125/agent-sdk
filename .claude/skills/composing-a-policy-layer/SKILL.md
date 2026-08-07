---
name: composing-a-policy-layer
description: Compose an external policy engine (allow or deny gating on an agent action) with the trustless-ai ERC stack without breaking recompute. Covers where a policy layer sits (beside the primitives, not above), what it must not swallow (identity, anchoring, and proof semantics stay in the ERCs), and what the decision must expose to stay independently checkable. Vendor neutral, with Bastion and CAPV as one worked example.
---

# Composing a policy layer

A policy engine answers one question about a proposed agent action: does it proceed or not. Amount caps, allowlists, rate limits, sanctions screening, a human in the loop gate. This is real and useful work, and it sits next to the ERC stack rather than inside it. This skill is about wiring one in without dissolving the guarantees the stack is built on.

The subject here is *any* external policy engine. Bastion and CAPV appear once, at the end, as a concrete worked example, not as the point.

## The one rule to internalize first

The verifier decides validity. The policy engine decides whether to proceed. These are different questions and they must stay in different components.

A policy layer that also starts deciding what an ERC record means, whether an inference proof is good, or whether an action should be anchored has stopped composing and started wrapping. Wrapping feels tidy because one object now speaks for the whole flow, but it inherits every breaking change in every ERC it speaks for, and it moves the trust boundary back onto itself. The whole reason the stack exists is that a reader does not have to trust any one component. A policy layer that swallows the primitives quietly undoes that.

## 1. Where a policy layer sits

Beside the primitives, on the path, gating the transition. Not above them, not translating them.

```text
ERC-8004 identity        proposed action        ERC-8263 anchor
    (who)          ->        (what)        ->      (recorded)
                               |
                         Policy Guard
                     (proceed or do not)
                               |
                       ERC-8274 verify
                       (inference proof)
```

The identity registry still says who the actor is. The verifier still says whether a proof is valid. The anchor still records the timeline. The policy Guard reads what it needs from those and returns a single verdict on whether this action is allowed to continue. It does not re-answer any of their questions.

A good test: remove the policy layer and the ERC flow still means exactly what it meant. Remove an ERC and the policy layer cannot silently keep asserting that ERC's guarantee on its own. If pulling the policy engine changes what the identity or the proof *means*, the layer was above, not beside.

## 2. What it must not swallow

Keep these in the ERCs. A policy engine that reimplements any of them has taken on a trust responsibility that was deliberately spread out.

- **Identity.** Who the actor is comes from ERC-8004. The policy engine may read the agent id and gate on it. It does not mint, rebind, or reinterpret identity. See `typescript/src/identity/ERC8004/README.md` for how far identity is fixed and where a caller has to stop trusting a convention.
- **Proof validity.** Whether an inference proof holds is ERC-8274's call, and it is one a caller can make themselves against the deployed, immutable verifier. See `typescript/src/verify/ERC8274/README.md`, including the split verdict there: the validity check is recompute-to-verify, while a separate derived digest is not. A policy layer must not collapse that distinction by re-asserting validity through its own endpoint.
- **Provenance.** The binding from raw input to what the model received is ERC-8299 (WYRIWE). A policy decision can depend on the provenance hash. It does not recompute provenance and present its own version as authoritative.
- **Anchoring.** The action timeline is ERC-8263. The policy verdict can be one input to what gets anchored. The policy engine does not decide the anchor's meaning.
- **Commitment semantics.** Observation and re-check commitments are ERC-8281 (OCP). The log is the ledger. A policy layer reuses those commitments, it does not redefine them.

The pattern across all five: the policy layer is a *consumer* of ERC facts and a *producer* of one new fact (the verdict). It is never the authority on facts the ERCs already own.

## 3. What the decision must expose to stay recomputable

This is the requirement most policy integrations miss. "Our engine evaluated the policy and it passed" is an assertion about a hosted endpoint. Nobody can check it. Per `INTEGRATIONS.md`, an integration whose claims are only assertable is still welcome, but it has to say so plainly, and a policy gate that only asserts is a weak link on a stack whose entire premise is independent verification.

The strong version emits the verdict as an artifact a third party can check without trusting the engine. Two shapes qualify.

**a) An on chain consumable verdict proof.** The engine proves, in zero knowledge, that the action clears a policy that is committed but never revealed, and a Guard contract consumes that proof on chain. Consuming reverts unless the verdict is for this domain, sits on an accepted policy root, burns an unused nullifier, and carries a valid proof. Now "the gate ran" is checkable, because the gate ran if and only if the Guard consumed the verdict. The policy itself stays private. The check does not.

**b) A recomputable commitment or signed record.** The engine emits an ERC-8281 style commitment, or a signed record, whose value a reader reproduces from public inputs. The membrane rule for this repo applies: ship a recompute or verify step and document it near the top of the README. If the decision digest is derivable, expose the derivation. Do not re-assert a number the reader could compute.

One discipline binds both shapes together with the rest of the SDK: if the policy layer needs to state something the SDK can already derive, it derives it rather than re-asserting it. Cross lane vectors live in `testkit/vectors/` and `recompute-kit/conformance/`, and an integration that computes the same values is welcome to check itself against them.

## 4. Worked example: Bastion and CAPV

Bastion is an external policy engine (pre-execution simulation, programmable rules, a human in the loop gate). On its own it would be a section 3 assertion: trust the endpoint that says the gate passed.

CAPV (Confidential Agent Policy Verdicts, draft ERC-8354) is the artifact that makes it checkable, and it is exactly shape (a). The engine evaluates the action against a secret policy and emits a zero knowledge verdict. An on chain Guard `consume`s it, checking domain, policy root, nullifier, and proof. The policy never appears on chain. The fact that a valid verdict was consumed does.

Crucially, this composes *beside* the primitives, matching sections 1 and 2. The Guard gates whether the action proceeds. It does not decide identity, it does not re-verify the inference proof, it does not redefine the anchor. In the composed run, the CAPV verdict lands through the same `IProofVerifier` boundary as an ERC-8274 inference proof and a `recompute` verdict, presented on one ledger and checkable against one socket. Two orthogonal guarantees on one action: the inference ran (ERC-8274, the recompute corner) and the resulting action cleared a policy that was never revealed (CAPV, the confidential corner).

Discussion and reference: the CAPV draft on Ethereum Magicians (`ethereum-magicians.org/t/draft-idea-confidential-agent-policy-verdicts/29088`), which builds on ERC-8004 and ERC-7812, and the integration itself at `github.com/zkos-labs/bastion`.

## Checklist for any policy integration

- The verifier still decides validity. The policy engine decides only proceed or not.
- Identity, proof, provenance, anchoring, and commitment semantics stay in their ERCs.
- The verdict is emitted as a checkable artifact (a consumed on chain proof, or a recomputable commitment), or the integration says plainly that its claim is only assertable.
- Anything the SDK can derive is derived, not re-asserted, and is checked against the conformance vectors where they exist.
