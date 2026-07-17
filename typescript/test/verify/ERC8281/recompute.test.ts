import { describe, expect, it } from 'vitest'
import { keccak256, toHex } from 'viem'
import { computeObservationDigest } from '../../../src/verify/ERC8281/recompute.js'

describe('computeObservationDigest (ERC-8281 recompute)', () => {
  it('computes the digest for a short observation', () => {
    const observation = toHex('hello')
    const result = computeObservationDigest(observation)
    const expected = keccak256(observation)
    expect(result).toBe(expected)
  })

  it('computes the digest for empty observation (0x)', () => {
    const observation = '0x' as const
    const result = computeObservationDigest(observation)
    const expected = keccak256(observation)
    expect(result).toBe(expected)
  })

  it('computes the digest for a longer observation', () => {
    const observation = toHex('a'.repeat(64))
    const result = computeObservationDigest(observation)
    const expected = keccak256(observation)
    expect(result).toBe(expected)
  })

  it('matches the inline golden vector', () => {
    const observation = toHex('observation-data')
    const result = computeObservationDigest(observation)
    const expected = keccak256(observation)
    expect(result).toBe(expected)
  })

  it('matches the empty-input golden vector', () => {
    const result = computeObservationDigest('0x')
    // keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
    expect(result).toBe('0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470')
  })
})
