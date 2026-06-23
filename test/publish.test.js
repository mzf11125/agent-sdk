'use strict';
// v0.2 tests — no network. Proves normalizeSpec kills checksum-case drift, and that publishCommit
// orchestrates the relay + OTS legs correctly over INJECTED I/O (refusing unsigned / malformed events).
const assert = require('assert');
const { schnorr } = require('@noble/curves/secp256k1');
const { sha256 } = require('@noble/hashes/sha256');
const {
  buildCommitEvent, artifactHash, normalizeSpec, publishCommit,
} = require('../src/index');

function hex(b) { return Array.from(b).map((x) => x.toString(16).padStart(2, '0')).join(''); }

const sk = sha256(new TextEncoder().encode('agent-sdk-v02-test-key'));
const pk = hex(schnorr.getPublicKey(sk));

// ---- normalizeSpec: the single anti-drift point ----------------------------------------------
const mixedCase = {
  job_id: 'job-1',
  target_wallet: '0xAbC0000000000000000000000000000000000001',
  output_address: '0xDEf0000000000000000000000000000000000002',
  asset_set: ['ENS', '0xFFf0000000000000000000000000000000000003'],
};
const lower = {
  job_id: 'job-1',
  target_wallet: '0xabc0000000000000000000000000000000000001',
  output_address: '0xdef0000000000000000000000000000000000002',
  asset_set: ['ENS', '0xfff0000000000000000000000000000000000003'],
};
// committer (checksum-cased) and escrow (lowercase) reach the SAME hash through normalizeSpec.
assert.strictEqual(
  artifactHash(normalizeSpec(mixedCase)),
  artifactHash(normalizeSpec(lower)),
  'normalizeSpec must make checksum-case and lowercase specs hash identically',
);
// non-address strings are left untouched (ENS label preserved, not lowercased oddly)
assert.deepStrictEqual(normalizeSpec({ name: 'MixedCaseLabel', n: 5 }), { name: 'MixedCaseLabel', n: 5 });
// idempotent
assert.strictEqual(artifactHash(normalizeSpec(normalizeSpec(mixedCase))), artifactHash(normalizeSpec(mixedCase)));

// ---- publishCommit: orchestrates injected I/O ------------------------------------------------
const { event } = buildCommitEvent({ spec: normalizeSpec(mixedCase), pubkey: pk, judgmentType: 'recovery_receipt' });

// unsigned → refuse (the SDK never signs)
assert.rejects(() => publishCommit({ event, relays: ['wss://r'] }), /signed/, 'unsigned event must be rejected');

// sign it
event.sig = hex(schnorr.sign(event.id, sk));

(async () => {
  // injected relay publisher: first relay OK, second fails, third throws → only the OK one counts
  const seen = [];
  const publishToRelay = async (relay, ev) => {
    seen.push(relay);
    assert.strictEqual(ev.id, event.id, 'publisher receives the exact event');
    if (relay === 'wss://bad') return false;
    if (relay === 'wss://throw') throw new Error('boom');
    return true;
  };
  let stampedId = null;
  const otsStamp = async (id) => { stampedId = id; return { ots_b64: 'deadbeef', pending: true }; };

  const res = await publishCommit({
    event,
    relays: ['wss://good', 'wss://bad', 'wss://throw'],
    publishToRelay,
    otsStamp,
  });
  assert.deepStrictEqual(seen, ['wss://good', 'wss://bad', 'wss://throw'], 'all relays attempted');
  assert.deepStrictEqual(res.relays, ['wss://good'], 'only OK relays recorded');
  assert.strictEqual(res.relay_count, 1);
  assert.strictEqual(res.anchored, true, 'anchored once at least one relay holds a copy');
  assert.strictEqual(stampedId, event.id, 'OTS stamps the event id');
  assert.deepStrictEqual(res.ots, { ots_b64: 'deadbeef', pending: true });
  assert.strictEqual(res.commitment_proof.event_id, event.id);
  assert.strictEqual(res.commitment_proof.mechanism, 'nostr-relay-publication + bitcoin-pow (opentimestamps)');

  // tampered id → refuse to anchor a malformed commit
  const bad = { ...event, content: event.content.replace('recovery_receipt', 'evil') };
  await assert.rejects(
    () => publishCommit({ event: bad, relays: [], publishToRelay }),
    /does not match/,
    'event whose id no longer matches its contents must be refused',
  );

  // no relays + no otsStamp → not anchored, ots null, never throws
  const empty = await publishCommit({ event, relays: [] });
  assert.strictEqual(empty.anchored, false);
  assert.strictEqual(empty.ots, null);

  console.log('ok — all agent-sdk v0.2 tests passed');
})();
