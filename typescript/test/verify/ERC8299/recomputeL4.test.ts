import { describe, expect, it } from 'vitest'
import { computeRawProposalHash, computeVerdictHash } from '../../../src/verify/ERC8299/recomputeL4.js'

// Golden vector cross-verified against invinoveritas's actual reference implementation
// (services/proof_signing.py, both artifact_hash and compute_decision_ref) on 2026-07-14 --
// same string in, byte-identical hash out, in both Python and this TypeScript port.
const ARTIFACT = 'test artifact content for cross-language verification'
const EXPECTED_RAW_PROPOSAL_HASH =
  '0xb8f70a237da212a272ecd09370acedbce6ca1d7df90745beafcac77e39697a88' as const

const VERDICT_FIELDS = {
  artifact_hash: 'b8f70a237da212a272ecd09370acedbce6ca1d7df90745beafcac77e39697a88',
  artifact_type: 'plan',
  policy_version: 'invinoveritas.review.v4',
  verdict: 'approve',
  source_class: 'agent_reported',
  vantage_limitation: null,
}
const PREIMAGE_FIELDS = [
  'artifact_hash',
  'artifact_type',
  'policy_version',
  'verdict',
  'source_class',
  'vantage_limitation',
] as const
const EXPECTED_VERDICT_HASH =
  'sha256:2970854c035d5aedb673b8523128665712895f62dd525c91fc8e858ad588ce58'

describe('ERC-8299 L4 recompute functions', () => {
  describe('computeRawProposalHash', () => {
    it('matches invinoveritas reference implementation golden vector', () => {
      expect(computeRawProposalHash(ARTIFACT)).toBe(EXPECTED_RAW_PROPOSAL_HASH)
    })

    it('handles empty artifact', () => {
      const result = computeRawProposalHash('')
      expect(result.startsWith('0x')).toBe(true)
      expect(result.length).toBe(66)
    })

    it('produces different hashes for different artifacts', () => {
      expect(computeRawProposalHash('a')).not.toBe(computeRawProposalHash('b'))
    })
  })

  describe('computeVerdictHash', () => {
    it('matches invinoveritas reference implementation golden vector', () => {
      expect(computeVerdictHash(VERDICT_FIELDS, PREIMAGE_FIELDS)).toBe(EXPECTED_VERDICT_HASH)
    })

    it('is order-independent in preimageFields (sorted internally, per JCS)', () => {
      const reversed = [...PREIMAGE_FIELDS].reverse()
      expect(computeVerdictHash(VERDICT_FIELDS, reversed)).toBe(EXPECTED_VERDICT_HASH)
    })

    it('treats a missing field the same as an explicit null', () => {
      const { vantage_limitation, ...withoutField } = VERDICT_FIELDS
      expect(computeVerdictHash(withoutField, PREIMAGE_FIELDS)).toBe(EXPECTED_VERDICT_HASH)
    })

    it('produces different hashes for different verdicts', () => {
      const rejected = { ...VERDICT_FIELDS, verdict: 'reject' }
      expect(computeVerdictHash(rejected, PREIMAGE_FIELDS)).not.toBe(EXPECTED_VERDICT_HASH)
    })
  })
})
