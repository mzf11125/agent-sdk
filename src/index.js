// @onchain-ai/agent-sdk — the off-chain verify/recompute layer for the onchain-ai boundary chain.
//
// One install, one API, to commit-before-outcome, anchor to public sources, and verify any layer's
// claims without trusting anyone. A thin, transparent wrapper over an auditable zero-I/O core — every
// convenience function here has a manual path you can step underneath (see README, both shown side by side).
//
// The trust anchor is the verify core, pinned: it is byte-compatible with the invinoveritas reference
// verifier `invinoveritas-verify` and with https://api.babyblueviper.com/verify-proof.

'use strict';
const { verifyProof, nostrEventId, PROOF_KIND } = require('./verify');
const { buildCommitEvent, artifactHash, canonical, normalizeSpec } = require('./commit');
const { verifyFullFlow } = require('./flow');
const { recompute } = require('./recompute');
const { publishCommit, relayPublish } = require('./publish');
const { packReceiptProof } = require('./onchain');

// The exact verify logic this SDK's one-liners run — so a one-liner is never a black box.
const CORE = { name: 'agent-sdk verify core', verifies_like: 'invinoveritas-verify@0.1.0', proof_kind: PROOF_KIND };

module.exports = {
  // build + hash + normalize (commit-before-outcome). normalizeSpec = the single anti-drift point.
  buildCommitEvent, artifactHash, canonical, normalizeSpec,
  // anchor to public sources (relay + OTS) — injected I/O, never signs
  publishCommit, relayPublish,
  // pack a signed event into the on-chain BIP340Verifier's receiptProof bytes (byte-aligned w/ verify)
  packReceiptProof,
  // verify (the trust anchor) + the composed gate
  verifyProof, nostrEventId, verifyFullFlow,
  // recompute a record from public events
  recompute,
  CORE, PROOF_KIND,
};
