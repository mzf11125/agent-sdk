// @onchain-ai/agent-sdk — commit.js
//
// Build a commit-before-outcome event. The commit IS a signed kind-30078 Nostr event whose content
// carries artifact_hash = a deterministic hash of the job spec. Because the spec (incl. the
// owner-bound output_address) is inside the hash, the commit is owner-bound by construction and
// non-portable. This module is pure: it constructs + hashes; signing and publishing are injected so
// the SDK never touches a private key and stays zero-I/O at its core.

'use strict';
const { sha256 } = require('@noble/hashes/sha256');
const { nostrEventId, PROOF_KIND } = require('./verify');

function toHex(bytes) {
  return Array.from(bytes).map((b) => b.toString(16).padStart(2, '0')).join('');
}

// Deterministic canonical JSON (sorted keys, recursive) so artifact_hash is reproducible by anyone.
function canonical(value) {
  if (Array.isArray(value)) return `[${value.map(canonical).join(',')}]`;
  if (value && typeof value === 'object') {
    return `{${Object.keys(value).sort().map((k) => `${JSON.stringify(k)}:${canonical(value[k])}`).join(',')}}`;
  }
  return JSON.stringify(value);
}

const EVM_ADDRESS = /^0x[0-9a-fA-F]{40}$/;

/**
 * The single normalization point (v0.2). Recursively lowercases EVM-address-shaped strings
 * (0x + 40 hex) anywhere in a spec, leaving everything else byte-identical. Run it on BOTH sides —
 * the committer (e.g. receipt.ts) and the escrow (expect_artifact_hash) — so artifact_hash is stable
 * regardless of input checksum-casing, and `canonical()` itself never has to special-case addresses.
 * This is what kills canonicalization drift: one function, identical on both sides.
 *
 *   const h      = artifactHash(normalizeSpec(spec));   // committer
 *   const expect = artifactHash(normalizeSpec(spec));   // escrow — byte-identical input → same hash
 */
function normalizeSpec(value) {
  if (Array.isArray(value)) return value.map(normalizeSpec);
  if (value && typeof value === 'object') {
    const out = {};
    for (const k of Object.keys(value)) out[k] = normalizeSpec(value[k]);
    return out;
  }
  if (typeof value === 'string' && EVM_ADDRESS.test(value)) return value.toLowerCase();
  return value;
}

/**
 * Hash a job spec to the value that goes in the commit AND that the escrow passes as expect_artifact_hash.
 * Put a job_id/salt in the spec so two legitimately-identical jobs stay distinct (and don't collide
 * under a nullifier). Keep result_ref / settled-tx OUT — that's the outcome leg, not the commit.
 * `artifactHash` is pure and case-agnostic about addresses — normalize upstream with `normalizeSpec`.
 */
function artifactHash(spec) {
  return toHex(sha256(new TextEncoder().encode(canonical(spec))));
}

/**
 * Construct the (unsigned) commit event. Caller signs it (inject your own signer / key management).
 * @param {object} p
 * @param {object} p.spec        e.g. { job_id, target_wallet, output_address, asset_set }
 * @param {string} p.pubkey      x-only hex of the signer
 * @param {string} p.judgmentType  e.g. "recovery_receipt" / "outcome_verifiable"
 * @param {string} [p.schema]    content.schema tag (default "onchain-ai.commit.v0")
 * @param {number} [p.createdAt] unix seconds; defaults to now. This is committed_at (pre-outcome).
 * @returns {{ event: object, id: string, artifact_hash: string }} event is unsigned (no `sig`).
 */
function buildCommitEvent({ spec, pubkey, judgmentType, schema = 'onchain-ai.commit.v0', createdAt }) {
  if (!spec || typeof spec !== 'object') throw new Error('spec must be an object');
  if (!pubkey) throw new Error('pubkey required');
  const created_at = Number(createdAt || Math.floor(Date.now() / 1000));
  const artifact_hash = artifactHash(spec);
  const content = JSON.stringify({
    schema,
    artifact_hash,
    committed_at: created_at,   // commit-before-outcome: the pre-outcome timestamp
    judgment_type: judgmentType,
  });
  const event = { pubkey: String(pubkey).toLowerCase(), created_at, kind: PROOF_KIND, tags: [], content };
  const id = nostrEventId(event);
  event.id = id;
  return { event, id, artifact_hash };
}

module.exports = { buildCommitEvent, artifactHash, canonical, normalizeSpec };
