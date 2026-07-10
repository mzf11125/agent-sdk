# ERC-8274 — AI Inference Proof Verification (Python)

**Recompute-to-verify: split verdict.** YES for the core proof-validity claim, NO for independently confirming a third party's reported `verification_digest`. Same verdict as the TypeScript implementation — see its README for the full rationale (summary below).

## Three contracts, not one

```
Settlement Contract (IAgentVerifiable) — declares which IAgentVerifier it trusts
    └── IAgentVerifier — stateful: agent authorization + proof routing
            └── IProofVerifier — stateless: raw cryptographic proof check (zkML/opML/TEE)
```

## What IS recompute-to-verify: the core proof-validity claim

`IProofVerifier.verify()` is a deterministic function of public inputs, deployed as immutable bytecode anyone can call directly — no special permission, no trust in whoever originally submitted the claim required. Calling it yourself and getting back `valid` is genuine independent recomputation: the same trust model as re-checking a Merkle proof or a signature yourself. `AgentVerifierClient.verify()`'s `valid` field inherits this, since it routes internally to `IProofVerifier.verify()`. This is why `ProofVerifierClient.verify()` is a plain read-only `.call()` rather than a permissioned action — it **is** this ERC's recompute-to-verify mechanism for its central claim.

## What is NOT recompute-to-verify: the reported verification_digest

`IAgentVerifier.verify()`'s digest formula (`keccak256(abi.encode(taskId, agentId, inputHash, outputHash, valid, agentProofProfile))`) references `agentProofProfile`, which is not a function parameter, not an event field, and not exposed by any getter in the interface — a caller cannot independently recompute the digest from public data alone to catch a third party misreporting it. This is a narrower limitation than the core claim above: it's about the digest's use as an audit-trail artifact, not whether the underlying proof-validity result can be checked.

If `agent-ercs` fixes the `agentProofProfile` derivation and exposes it, this second gap goes away — revisit via `/update-erc`.

## API

- `ProofVerifierClient` (`client.py`) — wraps `IProofVerifier`: `verify(input_hash, output_hash, metadata, proof)` (read-only `.call()`, no gas needed — this is the recompute-to-verify call), `proof_system()`, `proof_profile()`.
- `AgentVerifierClient` (`client.py`) — wraps `IAgentVerifier`: `verify(task_id, agent_id, input_hash, output_hash, proof) → (valid, verification_digest)` (broadcasts, since it records a `VerificationCompleted` event on-chain).
- `get_trusted_verifier(rpc_url, settlement_address)` (`client.py`) — a standalone function, not a class, since `IAgentVerifiable` is a single getter.

Tests deploy `testkit`'s `MockProofVerifier`, `MockAgentVerifier`, and `MockAgentVerifiable` (reference implementations for local testing only) to a local `anvil` node and call through these clients.
