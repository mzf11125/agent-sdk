# ERC-8263 — OnChain Proof Anchor (TypeScript)

**Recompute-to-verify: NO.**

ERC-8263 is a write-side anchor floor. The contract emits `AnchorProof` events with a non-zero `proofHash` commitment together with an identity-scheme byte and a 32-byte `agentId`. It performs no verification and no profile detection — `proofHash` is explicitly "an opaque non-zero bytes32 commitment." The ERC's own README states that "the contract does not constrain the hash algorithm or canonicalization producing proofHash; profile-level rules apply at the commitment layer."

Since the ERC does **not** fix how `proofHash` is computed, a generic SDK cannot independently recompute it. Verification logic exists at the profile/application layer, not the ERC layer, so no `verify()` method is generated here.

## API

- `anchor(agentIdScheme, agentId, proofHash)` — emit an `AnchorProof` with empty `aux`.
- `anchorWithAux(agentIdScheme, agentId, proofHash, aux)` — emit an `AnchorProof` with opaque extension bytes.

## Layer 2 — Pure recompute: NONE

The ERC defines no deterministic computation that can be reproduced off-chain from public inputs. No conformance vectors exist in `recompute-kit` for ERC-8263.

## Tests

Tests deploy `testkit`'s `MockOnChainProof` (a reference implementation for local testing only — see `testkit/contracts/mocks/anchor/ERC8263/`) to a local `anvil` node and call through this client. The mock enforces the canonical-form guards (`proofHash != 0`, scheme validation) specified in the ERC.
