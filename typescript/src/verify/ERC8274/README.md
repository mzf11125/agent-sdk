# ERC-8274 — AI Inference Proof Verification (TypeScript)

**Recompute-to-verify: split verdict.** YES for the core proof-validity claim, NO for independently confirming a third party's reported `verificationDigest`. These are two separable claims within the same ERC — see below.

## Three contracts, not one

Unlike ERC-8004, this ERC is three separate interfaces meant to be deployed as three separate contracts that reference each other by address:

```
Settlement Contract (IAgentVerifiable) — declares which IAgentVerifier it trusts
    └── IAgentVerifier — stateful: agent authorization + proof routing
            └── IProofVerifier — stateless: raw cryptographic proof check (zkML/opML/TEE)
```

## What IS recompute-to-verify: the core proof-validity claim

`IProofVerifier.verify()` is a deterministic function of public inputs (`inputHash`, `outputHash`, `metadata`, `proof`), deployed as immutable bytecode that anyone can call directly — no special permission, no trust in whoever originally submitted the claim required. Calling it yourself and getting back `valid` genuinely is independent recomputation: the same trust model as re-checking a Merkle proof or a signature yourself — you trust the (public, auditable, immutable) checking procedure once, not each party who happens to invoke it. `AgentVerifierClient.verify()`'s `valid` field inherits this same property, since it routes internally to `IProofVerifier.verify()`.

This is why `ProofVerifierClient.verify()` is implemented as a free, read-only simulated call rather than treated as an opaque, permissioned action — it **is** this ERC's recompute-to-verify mechanism for its central claim. (The ERC being "proof-system-agnostic" — zkML, opML, TEE all differ — doesn't undermine this: the SDK doesn't need to know or reimplement the specific cryptographic algorithm off-chain, because the deployed, immutable contract already **is** the public, callable-by-anyone checking procedure.)

## What is NOT recompute-to-verify: the reported verificationDigest

`IAgentVerifier.verify()`'s own doc comment claims the emitted `VerificationCompleted` event "carries the full preimage fields for OCP recompute → compare → confirm," and defines `verificationDigest = keccak256(abi.encode(taskId, agentId, inputHash, outputHash, valid, agentProofProfile))`. But **`agentProofProfile` is not a parameter of `verify()`, not a field of `VerificationCompleted`, and not exposed by any getter anywhere in `IAgentVerifier`.** A caller who only has the standard interface plus the emitted event is missing one of the six digest preimage components, and so cannot independently recompute `verificationDigest` to catch a third party (an indexer, a UI, another contract) misreporting it. This is a narrower limitation than the core claim above: it affects only the digest's use as an audit-trail/tamper-detection artifact, not whether the underlying proof-validity result can be checked — that part is covered above.

If `agent-ercs` fixes the `agentProofProfile` derivation (e.g. requires it to equal `IProofVerifier.proofProfile()` and exposes that on `IAgentVerifier`), this second gap goes away — revisit via `/update-erc`.

## API

- `ProofVerifierClient` (`proofVerifierClient.ts`) — wraps `IProofVerifier`: `verify(inputHash, outputHash, metadata, proof)` (read-only, no gas needed — this is the recompute-to-verify call), `proofSystem()`, `proofProfile()`.
- `AgentVerifierClient` (`agentVerifierClient.ts`) — wraps `IAgentVerifier`: `verify(taskId, agentId, inputHash, outputHash, proof) → { valid, verificationDigest }` (broadcasts, since it records a `VerificationCompleted` event on-chain).
- `getTrustedVerifier(config)` (`agentVerifiable.ts`) — a standalone function, not a class, since `IAgentVerifiable` is a single getter: reads the `agentVerifier()` address a settlement contract trusts.

Tests deploy `testkit`'s `MockProofVerifier`, `MockAgentVerifier`, and `MockAgentVerifiable` (reference implementations for local testing only — `MockProofVerifier`'s "proof" is a simple hash check, not real zkML/opML/TEE cryptography) to a local `anvil` node and call through these clients.
