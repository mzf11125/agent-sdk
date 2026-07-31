import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { ReviewGateClient } from '../../../src/governance/InvinoVeritas/client.js'
import type { ReviewResponse } from '../../../src/governance/InvinoVeritas/types.js'

// ── Golden vector ──────────────────────────────────────────────────────────
// A real, live signed proof pulled from api.babyblueviper.com/review
// (artifact: "rm -rf /var/log/nginx/*.log", verdict: reject, confidence
// 0.98), not a hand-built fixture — the same "don't fabricate a plausible-
// looking event, use a real one" discipline invinoveritas's own test suite
// runs on itself.
const GOLDEN_EVENT = {
  id: '17a6dc4b8caf56f8158d236a756f572aa34d0ab3d56d747ae358192f2e7c4dbc',
  pubkey: '6786e18a864893a900bd9858e650f67ccc3513f248fed374b591e2ff6922fbb7',
  created_at: 1785517683,
  kind: 30078,
  tags: [
    ['d', 'invinoveritas-proof-c265b3e2718868b414aece34577abb48abd3fccbed5ad7a8a1aa9abef878b78a-1785517683-7e45b4f5'],
    ['t', 'invinoveritas'],
    ['t', 'proof'],
    ['schema', 'invinoveritas.verdict_proof.v1'],
  ],
  content:
    '{"artifact_hash":"c265b3e2718868b414aece34577abb48abd3fccbed5ad7a8a1aa9abef878b78a","artifact_type":"shell_command","confidence":0.98,"conformance_suite":"https://github.com/babyblueviper1/preaction-governance-conformance","decision_ref":"sha256:fe334a7af96677208b39ebf5aeaa106c7eebd7af141fd547949efe5a32c2ec0d","decision_ref_preimage_fields":["artifact_hash","artifact_type","policy_version","verdict","source_class","vantage_limitation"],"decision_ref_preimage_rule":"every name in decision_ref_preimage_fields is a key in the hashed preimage object, always -- absent fields (e.g. vantage_limitation when not applicable) are present as JSON null, never omitted from the object.","independent_nodes":["https://invinoveritas-castra.babyblueviper.workers.dev/verify","https://babyblueviper1--2aba75da693711f185891607ee4eb77e.web.val.run"],"key_id":"6786e18a864893a900bd9858e650f67ccc3513f248fed374b591e2ff6922fbb7","platform":"invinoveritas","policy_version":"invinoveritas.review.v5","schema":"invinoveritas.verdict_proof.v1","source_class":"agent_reported","summary_hash":"3bef8fa561c538f34aa16620b06b029c4a6b82135b94444b29684a256480fa2a","verdict":"reject","verified_at":1785517683,"verifier_keys":"https://api.babyblueviper.com/.well-known/verifier-keys.json","verifier_pubkey":"6786e18a864893a900bd9858e650f67ccc3513f248fed374b591e2ff6922fbb7","verify_how":"Easiest: install the offline verifier above and recompute locally. Or POST this proof\'s signed `event` to verify_url (or an independent_node), OR run NIP-01 yourself: recompute the Nostr event id = sha256([0,pubkey,created_at,kind,tags,content]), verify the schnorr signature against verifier_pubkey. valid ⇒ invinoveritas issued this, untampered. No trust required.","verify_offline":"npm i invinoveritas-verify  ·  pip install invinoveritas-verify  — recompute this proof on your own machine against verifier_pubkey; you never have to call us.","verify_url":"https://api.babyblueviper.com/verify-proof"}',
  sig: 'ce2d72387fa9d8177dd9a247009de5c21a06cba2a823c6a816b24297a5df238065351f1cf8837113dfe84ca5eab4750adef9e4bc8c3b895bea2af51e82449ca1',
}

function makeResponse(overrides: Partial<ReviewResponse> = {}): ReviewResponse {
  return {
    status: 'success',
    verdict: 'reject',
    confidence: 0.98,
    summary: 'This irreversibly deletes all Nginx log files.',
    issues: [],
    proof: {
      proof_payload: {},
      signature_type: 'nostr_event',
      event: GOLDEN_EVENT,
    },
    ...overrides,
  }
}

describe('ReviewGateClient.verifyLocal — pure offline recompute', () => {
  const client = new ReviewGateClient({ apiKey: 'unused-for-verify-tests' })

  it('confirms a genuine, untampered proof (real golden vector)', () => {
    const result = client.verifyLocal(makeResponse())
    expect(result.valid).toBe(true)
    expect(result.checks.id_integrity).toBe(true)
    expect(result.checks.signature_valid).toBe(true)
    expect(result.checks.issued_by_invinoveritas).toBe(true)
    expect(result.checks.is_proof_event).toBe(true)
  })

  it('rejects a single-byte-tampered content field (signature no longer matches)', () => {
    const tampered = {
      ...GOLDEN_EVENT,
      content: GOLDEN_EVENT.content.replace('"confidence":0.98', '"confidence":0.99'),
    }
    const result = client.verifyLocal(makeResponse({ proof: { proof_payload: {}, signature_type: 'nostr_event', event: tampered } }))
    expect(result.valid).toBe(false)
    // The recomputed event id no longer matches the claimed id once content changes,
    // so id_integrity fails first — exactly the tamper-evidence property this exists for.
    expect(result.checks.id_integrity).toBe(false)
  })

  it('rejects a proof signed by a different key (not invinoveritas)', () => {
    const wrongPubkey = {
      ...GOLDEN_EVENT,
      pubkey: '0'.repeat(64),
    }
    const result = client.verifyLocal(makeResponse({ proof: { proof_payload: {}, signature_type: 'nostr_event', event: wrongPubkey } }))
    expect(result.valid).toBe(false)
  })

  it('throws a clear error when the response has no proof.event (sign:false was used)', () => {
    const client2 = new ReviewGateClient({ apiKey: 'unused' })
    expect(() => client2.verifyLocal(makeResponse({ proof: undefined }))).toThrow(/sign:true/)
  })
})

describe('ReviewGateClient.review — HTTP contract, no live network call', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
  })

  it('POSTs the expected body and headers to /review', async () => {
    const fetchMock = vi.fn(async (url: string | URL, init?: RequestInit) => {
      expect(String(url)).toBe('https://api.babyblueviper.com/review')
      expect(init?.method).toBe('POST')
      expect((init?.headers as Record<string, string>).Authorization).toBe('Bearer test-key')
      const parsedBody = JSON.parse(String(init?.body))
      expect(parsedBody.artifact).toBe('rm -rf /tmp/*')
      expect(parsedBody.artifact_type).toBe('shell_command')
      expect(parsedBody.sign).toBe(true)
      return new Response(JSON.stringify(makeResponse()), { status: 200 })
    })
    globalThis.fetch = fetchMock as unknown as typeof fetch

    const client = new ReviewGateClient({ apiKey: 'test-key' })
    const result = await client.review('rm -rf /tmp/*', { artifactType: 'shell_command' })

    expect(fetchMock).toHaveBeenCalledOnce()
    expect(result.verdict).toBe('reject')
  })

  it('defaults sign to true when not specified', async () => {
    const fetchMock = vi.fn(async (_url: string | URL, init?: RequestInit) => {
      const parsedBody = JSON.parse(String(init?.body))
      expect(parsedBody.sign).toBe(true)
      return new Response(JSON.stringify(makeResponse()), { status: 200 })
    })
    globalThis.fetch = fetchMock as unknown as typeof fetch

    const client = new ReviewGateClient({ apiKey: 'test-key' })
    await client.review('some action')
    expect(fetchMock).toHaveBeenCalledOnce()
  })

  it('honors a custom baseUrl', async () => {
    const fetchMock = vi.fn(async (url: string | URL) => {
      expect(String(url)).toBe('https://staging.example.com/review')
      return new Response(JSON.stringify(makeResponse()), { status: 200 })
    })
    globalThis.fetch = fetchMock as unknown as typeof fetch

    const client = new ReviewGateClient({ apiKey: 'test-key', baseUrl: 'https://staging.example.com/' })
    await client.review('some action')
    expect(fetchMock).toHaveBeenCalledOnce()
  })

  it('throws with response detail on a non-2xx status', async () => {
    const fetchMock = vi.fn(async () => new Response('insufficient balance', { status: 402 }))
    globalThis.fetch = fetchMock as unknown as typeof fetch

    const client = new ReviewGateClient({ apiKey: 'test-key' })
    await expect(client.review('some action')).rejects.toThrow(/402/)
  })
})
