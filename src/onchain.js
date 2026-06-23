// @trustless-ai/agent-sdk — onchain.js (v0.2)
//
// packReceiptProof: pack a SIGNED commit event into the exact `receiptProof` bytes the on-chain
// `BIP340Verifier` (IReceiptVerifier impl A) consumes. Same anti-drift discipline as normalizeSpec —
// off-chain `verifyFullFlow()` and on-chain `verify()` read byte-identical input, because BOTH derive
// the message id as sha256 of the SAME NIP-01 preimage produced here.
//
// Layout = Solidity `abi.encode(bytes32 px, bytes32 rx, bytes32 s, bytes preimage)`:
//   word0 px              (pubkey x-only)
//   word1 rx             (signature R.x  = sig[0:32])
//   word2 s              (signature s    = sig[32:64])
//   word3 0x80           (offset to the bytes tail)
//   word4 len(preimage)
//   ...   preimage, right-padded to a 32-byte multiple
// where preimage = JSON.stringify([0, pubkey, created_at, kind, tags, content]) — exactly the bytes
// `nostrEventId` hashes. Zero-dependency (no ethers/web3): a hand-rolled abi.encode for this one shape.

'use strict';

function _hexToBytes(hex) {
  const h = hex.startsWith('0x') ? hex.slice(2) : hex;
  if (h.length % 2 !== 0) throw new Error('odd-length hex');
  const out = new Uint8Array(h.length / 2);
  for (let i = 0; i < out.length; i++) out[i] = parseInt(h.substr(i * 2, 2), 16);
  return out;
}

function _bytesToHex(bytes) {
  let s = '';
  for (const b of bytes) s += b.toString(16).padStart(2, '0');
  return s;
}

// left-pad a 32-byte word (hex, no 0x) to 64 chars
function _word(hex32) {
  const h = (hex32.startsWith('0x') ? hex32.slice(2) : hex32).toLowerCase();
  if (h.length > 64) throw new Error('word too large');
  return h.padStart(64, '0');
}

// uint256 -> 32-byte word hex
function _uintWord(n) {
  return BigInt(n).toString(16).padStart(64, '0');
}

/**
 * @param {object} event  a SIGNED kind-30078 event { pubkey, created_at, kind, tags, content, sig }.
 *   `sig` is the 64-byte BIP-340 signature (hex): rx (first 32) ‖ s (last 32).
 * @returns {string} 0x-prefixed hex of abi.encode(px, rx, s, preimage) — feed straight to verify().
 */
function packReceiptProof(event) {
  if (!event || typeof event !== 'object') throw new Error('event must be an object');
  if (!event.sig) throw new Error('event must be signed (needs sig) before packing');
  const sig = (event.sig.startsWith('0x') ? event.sig.slice(2) : event.sig).toLowerCase();
  if (sig.length !== 128) throw new Error('sig must be 64 bytes (128 hex chars)');

  const px = _word(String(event.pubkey));
  const rx = _word(sig.slice(0, 64));
  const s = _word(sig.slice(64, 128));

  const preimageStr = JSON.stringify([
    0, String(event.pubkey).toLowerCase(), Number(event.created_at), Number(event.kind),
    event.tags || [], String(event.content),
  ]);
  const preimage = new TextEncoder().encode(preimageStr);
  const preimageHex = _bytesToHex(preimage);
  const padded = preimageHex.padEnd(Math.ceil(preimageHex.length / 64) * 64, '0');

  const head = px + rx + s + _uintWord(0x80);     // 4 words; offset to bytes tail = 0x80
  const tail = _uintWord(preimage.length) + padded; // length word + padded data
  return '0x' + head + tail;
}

module.exports = { packReceiptProof, _hexToBytes, _bytesToHex };
