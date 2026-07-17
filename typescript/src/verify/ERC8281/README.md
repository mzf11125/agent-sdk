# ERC-8281 — Observation Commitment Protocol (OCP)

Client bindings for `IObservationCommitment` — an on-chain commitment anchor.
Anchors an opaque `digest` on-chain via `record(bytes32)`, emitting
`Recorded(digest, committer)` as tamper-evident proof that an observation was
committed to at a specific block, without revealing the observation itself.

## API

### `ObservationCommitmentClient`

Requires an `Account` for transaction signing.

| Method | Description | State |
| --- | --- | --- |
| `record(digest)` | Commit a digest on-chain, returns `TransactionReceipt` | write |
| `parseRecordedEvent(receipt)` | Extract the `Recorded` event from a receipt | read |
| `supportsObservationCommitment()` | ERC-165 check for `0xb5c645bd` | read |

There is no on-chain getter — the event log IS the ledger.

## Layer-2 recompute

`computeObservationDigest(observation: Hex): Hex` — pure function that
reproduces `keccak256(observation)`, the core OCP commitment step. Verified
against golden conformance vectors.

## Verification

Verification is off-chain and recompute-based: a verifier re-derives the
digest from the primary artifact and confirms the matching `Recorded` log
exists at the claimed chain/block position. The proof envelope pins
`chain_id` and the receipt log position for unambiguous selection.
