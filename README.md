# @onchain-ai/agent-sdk

The off-chain **verify / recompute** layer of the [onchain-ai](https://github.com/onchain-ai) boundary
chain. One install to **commit-before-outcome**, anchor to public sources, and **verify any layer's
claims trusting no one**.

```bash
npm install @onchain-ai/agent-sdk
```

> **Status: v0.2.** The pure core (verify, commit-hash, `normalizeSpec`, the full-flow gate, recompute)
> is implemented and tested. v0.2 adds the **anchor leg** — `publishCommit()` (relay publish + OTS) —
> as an injected-I/O convenience over the zero-I/O core (see "Why I/O is injected" below), plus
> `normalizeSpec()`, the single anti-drift normalization point both committer and escrow share. First
> consumer: [`hack-ens-recovery`](https://github.com/TMerlini/hack-ens-recovery).

## Recompute / verify — run it yourself

```bash
npm test     # build a commit, sign it, verify it, exercise the gate — no network
```

## The trust anchor, and the convenience layer over it (both shown, never a black box)

Per the org [CONTRIBUTING](https://github.com/onchain-ai/.github/blob/main/CONTRIBUTING.md), the
zero-I/O verify core is first-class and you can always step underneath the one-liner:

```js
const { verifyFullFlow, verifyProof } = require('@onchain-ai/agent-sdk');

// convenience: the whole gate in one call
const gate = verifyFullFlow({
  proofEvent, expectArtifactHash, expectPubkey,
  schemaPrefix: 'onchain-ai.', relaySeen, otsVerified,
});
// gate.ok === verify.valid && artifact_hash_matches && anchored

// underneath: the exact same trust anchor, by hand — recompute the NIP-01 id + BIP-340 sig yourself
const v = verifyProof(proofEvent, { expectPubkey });   // { valid, checks, proof_payload }
```

`verifyProof` is byte-compatible with the [`invinoveritas-verify`](https://www.npmjs.com/package/invinoveritas-verify)
reference verifier and with `https://api.babyblueviper.com/verify-proof`. `CORE.verifies_like` pins the
exact core logic version your one-liner runs.

## The gate (never `valid` alone)

```
ok = valid  AND  artifact_hash_matches  AND  anchored(relaySeen && otsVerified)
```

`valid` only proves the receipt is a genuine signed proof — **not** that it is *this* job's, and proofs
carry no nonce/expiry so a valid receipt is **replayable**. So a consumer (e.g. an escrow) must gate on
all three **plus** an on-chain delivery check (assets actually landed at `output_address`) **plus** a
nullifier (mark the `artifact_hash` spent on release). The SDK does the off-chain three; the on-chain
two are the contract's job.

## Commit-before-outcome

```js
const { buildCommitEvent, artifactHash, normalizeSpec } = require('@onchain-ai/agent-sdk');

// artifact_hash = canonical hash of the job spec. Put a job_id/salt in so two identical jobs stay
// distinct; keep result_ref / settled-tx OUT (that's the outcome leg). normalizeSpec FIRST (below).
const spec = normalizeSpec({ job_id, target_wallet, output_address, asset_set });
const { event, artifact_hash } = buildCommitEvent({ spec, pubkey, judgmentType: 'recovery_receipt' });
event.sig = yourSigner(event.id);   // signing/keys stay yours — the SDK never touches a private key
```

`committed_at` is set to the event's `created_at`, so the commit provably predates its outcome once
anchored (relay copy + Bitcoin PoW via OpenTimestamps, `ots verify -d <event_id>`). The matching
read shape is `GET /ledger/{entry}/commitment`; the outcome leg is `GET /ledger/{entry}/outcome`.

## `normalizeSpec` — one function, identical on both sides (kills canonicalization drift)

The committer and the escrow must hash **byte-identical** input, or `artifact_hash_matches` silently
fails. `artifactHash` is pure and **case-agnostic about addresses by design** — so normalize *upstream*,
at a single shared point, rather than special-casing inside `canonical()`:

```js
const h      = artifactHash(normalizeSpec(spec));   // committer (receipt.ts)
const expect = artifactHash(normalizeSpec(spec));   // escrow — same fn, same input → same hash
```

`normalizeSpec` recursively lowercases EVM-address-shaped strings (`0x` + 40 hex) anywhere in the spec
and leaves everything else untouched. Run it on both sides and checksum-casing can never cause a miss.

## Anchor — `publishCommit()` (relay + OTS), injected I/O, never signs

```js
const { publishCommit } = require('@onchain-ai/agent-sdk');

// event must already be SIGNED — the SDK never holds a key.
const res = await publishCommit({
  event,
  relays: ['wss://relay.damus.io', 'wss://nos.lol'],
  // defaults: relays via the runtime's WebSocket (Node>=21/Bun/browser); inject your own on older Node.
  otsStamp: async (id) => myOtsClient.stamp(id),   // injected: needs the `ots` calendar I/O
});
// → { event_id, published_at, relays, relay_count, ots, anchored, commitment_proof }
// commitment_proof is the exact shape GET /ledger/{entry}/commitment mirrors, so verifyFullFlow()
// and /ledger read the same thing. The Bitcoin-PoW leg confirms later: `ots verify -d <event_id>`.
```

There is **no `POST /ledger/commit`** and there shouldn't be — `/ledger` is read-only; routing the
anchor through a server would re-centralize the exact thing the commitment-proof model makes trustless.
`publishCommit` anchors to public sources anyone re-derives; nothing routes through us.

## Why I/O is injected

Relay fetch and OTS calendar access are network, environment, and policy dependent. Keeping them out of
the core means the trust anchor is pure, deterministic, and testable, and you bring your own relay/OTS
client. `publishCommit()` is the thin convenience layer over that core — every leg of it is injectable.

## License

Apache-2.0.
