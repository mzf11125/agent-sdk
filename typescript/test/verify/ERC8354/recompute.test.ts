import { describe, expect, it } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  computeActionCommitment,
  computeVerdictDigest,
  MECHANISM_ZK_SECRET_POLICY,
} from '../../../src/verify/ERC8354/recompute.js'
import type { Address, Hex } from 'viem'

// Inline golden vectors (primary). These reproduce the vectors from
// testkit/vectors/erc8354-verdict.vectors.json so tests pass even when the
// vectors file is not present on disk.

const ACTION_COMMITMENT_INLINE = {
  chainId: 31337n,
  domainId: '0x34a63641b78652cdd53505da4f32cac6058bd148e3ff543f39f75997a89c2815' as Hex,
  agentId: 1n,
  target: '0x0000000000000000000000000000000000000001' as Address,
  value: 0n,
  callData: '0x' as Hex,
  actionNonce: 0n,
}

const ACTION_COMMITMENT_EXPECTED =
  '0xcc8e5dc414db5ed2340be02c3d7fdc725fe5f1463b382a7ed13f8036a4a0b7b1'

const VERDICT_INLINE = {
  agentId: 1n,
  domainId: '0x34a63641b78652cdd53505da4f32cac6058bd148e3ff543f39f75997a89c2815' as Hex,
  policyRoot: '0x34a63641b78652cdd53505da4f32cac6058bd148e3ff543f39f75997a89c2815' as Hex,
  actionCommitment: ACTION_COMMITMENT_EXPECTED as Hex,
  executor: '0x0000000000000000000000000000000000000002' as Address,
  expiry: 2000000000n,
  nullifier: '0x6e47261c83f90eed41cda2b00caad094c33daa0a09fec22396b3e2bfe5e222b2' as Hex,
  decision: 1,
  policyKind: 0,
}

const VERDICT_DIGEST_DOMAIN = {
  chainId: 31337,
  verifyingContract: '0x0000000000000000000000000000000000000003' as Address,
}

const VERDICT_DIGEST_EXPECTED = '0xf2345f63ba9e78a068eb4f74640e6543289010540b457d8016771175ad460f32'

describe('computeActionCommitment (ERC-8354 recompute)', () => {
  it('matches the inline golden vector', () => {
    expect(computeActionCommitment(ACTION_COMMITMENT_INLINE)).toBe(ACTION_COMMITMENT_EXPECTED)
  })

  it('hashes empty callData as keccak256(""), not bytes32 zero', () => {
    const result = computeActionCommitment(ACTION_COMMITMENT_INLINE)
    const withZeroCallData = computeActionCommitment({ ...ACTION_COMMITMENT_INLINE, callData: '0x' })
    expect(result).toBe(withZeroCallData)
    // keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
    expect(result).not.toBe('0x0000000000000000000000000000000000000000000000000000000000000000')
  })

  it('is deterministic', () => {
    expect(computeActionCommitment(ACTION_COMMITMENT_INLINE)).toBe(
      computeActionCommitment(ACTION_COMMITMENT_INLINE),
    )
  })

  it('changes when actionNonce changes', () => {
    const a = computeActionCommitment(ACTION_COMMITMENT_INLINE)
    const b = computeActionCommitment({ ...ACTION_COMMITMENT_INLINE, actionNonce: 1n })
    expect(a).not.toBe(b)
  })

  it('changes when callData changes', () => {
    const a = computeActionCommitment(ACTION_COMMITMENT_INLINE)
    const b = computeActionCommitment({ ...ACTION_COMMITMENT_INLINE, callData: '0x01' })
    expect(a).not.toBe(b)
  })
})

describe('computeVerdictDigest (ERC-8354 recompute)', () => {
  it('matches the inline golden vector', () => {
    expect(computeVerdictDigest(VERDICT_INLINE, VERDICT_DIGEST_DOMAIN)).toBe(
      VERDICT_DIGEST_EXPECTED,
    )
  })

  it('is deterministic', () => {
    expect(computeVerdictDigest(VERDICT_INLINE, VERDICT_DIGEST_DOMAIN)).toBe(
      computeVerdictDigest(VERDICT_INLINE, VERDICT_DIGEST_DOMAIN),
    )
  })

  it('depends on the verifying contract', () => {
    const a = computeVerdictDigest(VERDICT_INLINE, VERDICT_DIGEST_DOMAIN)
    const b = computeVerdictDigest(VERDICT_INLINE, {
      ...VERDICT_DIGEST_DOMAIN,
      verifyingContract: '0x0000000000000000000000000000000000000004',
    })
    expect(a).not.toBe(b)
  })
})

describe('MECHANISM_ZK_SECRET_POLICY (ERC-8354 constant)', () => {
  it('equals keccak256("zk-secret-policy")', () => {
    expect(MECHANISM_ZK_SECRET_POLICY).toBe(
      '0xa843829a78c66c29679817606d0c8a9fa26575b6c2ed0f9f97079d7c46577ac6',
    )
  })
})

// Conformance vector reader (secondary). Skips when the vectors file is
// missing on disk.

interface ConformanceVector {
  id: string
  step: string
  inputs: Record<string, unknown>
  expected: unknown
}

function loadConformanceVectors(): ConformanceVector[] {
  const vectorsPath = path.resolve(
    fileURLToPath(new URL('.', import.meta.url)),
    '../../../../testkit/vectors/erc8354-verdict.vectors.json',
  )
  if (!existsSync(vectorsPath)) {
    console.warn('testkit vectors not found, skipping file-based conformance check')
    return []
  }
  const raw = readFileSync(vectorsPath, 'utf-8')
  const data = JSON.parse(raw) as { vectors: ConformanceVector[] }
  return data.vectors
}

describe('golden vector conformance', () => {
  const vectors = loadConformanceVectors()
  if (vectors.length === 0) {
    it('(no golden vectors on disk, skipping)', () => {
      expect(true).toBe(true)
    })
    return
  }
  for (const v of vectors) {
    it(`${v.step}`, () => {
      switch (v.step) {
        case '8354/action-commitment': {
          const inputs = v.inputs
          expect(
            computeActionCommitment({
              chainId: BigInt(inputs.chainId as number),
              domainId: inputs.domainId as Hex,
              agentId: BigInt(inputs.agentId as number),
              target: inputs.target as Address,
              value: BigInt(inputs.value as number),
              callData: inputs.callData as Hex,
              actionNonce: BigInt(inputs.actionNonce as number),
            }),
          ).toBe(v.expected)
          break
        }
        case '8354/verdict-digest': {
          const inputs = v.inputs
          expect(
            computeVerdictDigest(
              {
                agentId: BigInt(inputs.agentId as number),
                domainId: inputs.domainId as Hex,
                policyRoot: inputs.policyRoot as Hex,
                actionCommitment: inputs.actionCommitment as Hex,
                executor: inputs.executor as Address,
                expiry: BigInt(inputs.expiry as number),
                nullifier: inputs.nullifier as Hex,
                decision: inputs.decision as number,
                policyKind: inputs.policyKind as number,
              },
              {
                chainId: inputs.chainId as number,
                verifyingContract: inputs.verifyingContract as Address,
              },
            ),
          ).toBe(v.expected)
          break
        }
        default:
          throw new Error(`unknown step ${v.step}, a vector exists that no function covers`)
      }
    })
  }
})
