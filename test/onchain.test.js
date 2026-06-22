'use strict';
// v0.2 on-chain helper test — packReceiptProof produces the exact abi.encode(px,rx,s,preimage) bytes
// the BIP340Verifier consumes, and the embedded preimage hashes (sha256) to the event id the signature
// was made over. No network. (The Solidity side asserts the round-trip end-to-end in foundry.)
const assert = require('assert');
const { schnorr } = require('@noble/curves/secp256k1');
const { sha256 } = require('@noble/hashes/sha256');
const { buildCommitEvent, normalizeSpec, packReceiptProof, nostrEventId } = require('../src/index');

function hex(b) { return Array.from(b).map((x) => x.toString(16).padStart(2, '0')).join(''); }

const sk = sha256(new TextEncoder().encode('onchain-helper-test-key'));
const pk = hex(schnorr.getPublicKey(sk));
const spec = normalizeSpec({ job_id: 'j1', target_wallet: '0xAbC0000000000000000000000000000000000001',
  output_address: '0xDeF0000000000000000000000000000000000002', asset_set: ['x.eth'] });
const { event } = buildCommitEvent({ spec, pubkey: pk, judgmentType: 'recovery_receipt' });

// unsigned → must refuse (the SDK never signs)
assert.throws(() => packReceiptProof(event), /signed/, 'packing an unsigned event must throw');

event.sig = hex(schnorr.sign(event.id, sk));
const proof = packReceiptProof(event);

// shape: 0x + 4 head words (px,rx,s,offset) + length word + padded preimage
assert.ok(proof.startsWith('0x'), 'hex output');
const body = proof.slice(2);
assert.strictEqual(body.slice(0, 64), pk.padStart(64, '0'), 'word0 = px');
assert.strictEqual(body.slice(64, 128), event.sig.slice(0, 64), 'word1 = rx');
assert.strictEqual(body.slice(128, 192), event.sig.slice(64, 128), 'word2 = s');
assert.strictEqual(BigInt('0x' + body.slice(192, 256)), 0x80n, 'word3 = offset 0x80');

// the embedded preimage must hash to the event id the signature signed
const preLen = Number(BigInt('0x' + body.slice(256, 320)));
const preHex = body.slice(320, 320 + preLen * 2);
const preBytes = Uint8Array.from(preHex.match(/../g).map((h) => parseInt(h, 16)));
const preStr = new TextDecoder().decode(preBytes);
assert.strictEqual(hex(sha256(preBytes)), event.id, 'sha256(preimage) must equal the signed event id');
// and the preimage is the canonical NIP-01 serialization the verifier core also computes
assert.strictEqual(nostrEventId(event), event.id, 'event id is the canonical NIP-01 id');
assert.ok(preStr.startsWith(`[0,"${pk}",`), 'preimage is the NIP-01 array starting with [0,"<pubkey>",');

// padding: total body length is a whole number of 32-byte words
assert.strictEqual((body.length / 2) % 32, 0, 'abi tail padded to 32-byte multiple');

console.log('ok — all agent-sdk on-chain (packReceiptProof) tests passed');
