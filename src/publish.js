// @onchain-ai/agent-sdk — publish.js (v0.2)
//
// The anchor leg. A signed commit event becomes trustless when it is anchored to PUBLIC sources that
// anyone re-derives — without routing through any server (there is no POST /ledger/commit, and there
// shouldn't be: that would re-centralize the exact thing the commitment-proof model makes trustless).
//
// Two legs ship here, both wrapping the zero-dep scripts:
//   1. publish  the signed kind-30078 event to public Nostr relays → third-party copies + published_at
//   2. OTS-stamp the event id to Bitcoin → PoW precedence (`ots stamp`, later `ots verify -d <id>`)
//
// This is the SDK's I/O CONVENIENCE layer — explicitly NOT the zero-I/O core. It never signs (the
// event arrives signed) and every piece of I/O is injectable so it stays testable and auditable:
// pass your own `publishToRelay` / `otsStamp`, or use the provided defaults.

'use strict';
const { nostrEventId } = require('./verify');

/**
 * Default NIP-01 relay publisher over a WebSocket. Uses the runtime's global WebSocket
 * (Node >= 21, Bun, browsers). On older runtimes, inject your own `publishToRelay`
 * (e.g. one backed by the `ws` package) — that's the manual path this default sits on top of.
 * Resolves true only when the relay returns OK true for this exact event id.
 */
async function relayPublish(relayUrl, event, { timeoutMs = 8000 } = {}) {
  const WS = globalThis.WebSocket;
  if (!WS) throw new Error('no global WebSocket — inject publishToRelay (e.g. backed by the `ws` package)');
  return new Promise((resolve) => {
    let settled = false;
    const ws = new WS(relayUrl);
    const finish = (ok) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      try { ws.close(); } catch (_) { /* already closing */ }
      resolve(ok);
    };
    const timer = setTimeout(() => finish(false), timeoutMs);
    ws.onopen = () => { try { ws.send(JSON.stringify(['EVENT', event])); } catch (_) { finish(false); } };
    ws.onmessage = (m) => {
      try {
        const d = JSON.parse(typeof m.data === 'string' ? m.data : String(m.data));
        if (d[0] === 'OK' && d[1] === event.id) finish(d[2] === true);
      } catch (_) { /* ignore non-OK frames */ }
    };
    ws.onerror = () => finish(false);
  });
}

/**
 * Anchor a SIGNED commit event to public sources (relay publication + OTS-to-Bitcoin).
 *
 * Pure orchestration over injected I/O — the SDK never signs and never holds a key. Returns the
 * canonical commitment-proof shape that `GET /ledger/{entry}/commitment` mirrors, so verifyFullFlow()
 * and /ledger read the same thing. The OTS leg confirms later (Bitcoin needs a block) — `ots` returns
 * a proof now and you re-check it any time with `ots verify -d <event_id>`.
 *
 * @param {object}   p
 * @param {object}   p.event            signed kind-30078 event { id,pubkey,created_at,kind,tags,content,sig }
 * @param {string[]} [p.relays]         public relay URLs to publish to
 * @param {Function} [p.publishToRelay] async (relayUrl, event) => boolean. Default: `relayPublish`.
 * @param {Function} [p.otsStamp]       async (eventId) => proof. Injected (needs the `ots` calendar I/O).
 * @param {boolean}  [p.requireSig=true] refuse to publish an unsigned event (the SDK never signs).
 * @returns {Promise<{
 *   event_id: string, published_at: number, relays: string[], relay_count: number,
 *   ots: any, anchored: boolean, commitment_proof: object
 * }>}
 */
async function publishCommit({
  event,
  relays = [],
  publishToRelay = relayPublish,
  otsStamp,
  requireSig = true,
}) {
  if (!event || typeof event !== 'object') throw new Error('event must be an object');
  if (requireSig && !event.sig) {
    throw new Error('event must be signed before publishing — the SDK never signs (inject your own signer)');
  }
  // Integrity gate: never publish an event whose id doesn't match its own contents.
  if (event.id && nostrEventId(event).toLowerCase() !== String(event.id).toLowerCase()) {
    throw new Error('event.id does not match its contents — refusing to anchor a malformed commit');
  }

  const published = [];
  for (const relay of relays) {
    let ok = false;
    try { ok = await publishToRelay(relay, event); } catch (_) { ok = false; }
    if (ok) published.push(relay);
  }

  let ots = null;
  if (typeof otsStamp === 'function') {
    try { ots = await otsStamp(event.id); } catch (e) { ots = { error: String(e) }; }
  }

  const published_at = Math.floor(Date.now() / 1000);
  return {
    event_id: event.id,
    published_at,
    relays: published,
    relay_count: published.length,
    ots,
    // relay copies exist now; the Bitcoin-PoW leg confirms once the OTS proof upgrades (ots verify).
    anchored: published.length > 0,
    commitment_proof: {
      mechanism: 'nostr-relay-publication + bitcoin-pow (opentimestamps)',
      event_id: event.id,
      published_at,
      relays: published,
      how_to_check:
        'fetch the event by id from any listed relay, confirm created_at; ' +
        'verify the OTS proof to Bitcoin (ots verify -d <event_id>)',
    },
  };
}

module.exports = { publishCommit, relayPublish };
