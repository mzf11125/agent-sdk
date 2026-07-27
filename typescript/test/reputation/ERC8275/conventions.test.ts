import { describe, expect, it } from 'vitest'
import {
  governingConventionHash,
  pinWinRateBps,
  verifyWinRate,
  WIN_RATE_BPS_V0_HASH,
  WIN_RATE_BPS_V0_SPEC,
} from '../../../src/reputation/ERC8275/conventions.js'

const H = WIN_RATE_BPS_V0_HASH

describe('governing_convention_hash — win_rate.bps.v0', () => {
  it('reproduces the locked identity from the spec', () => {
    expect(governingConventionHash(WIN_RATE_BPS_V0_SPEC)).toBe(H)
  })
  it('pins at issuance', () => {
    expect(pinWinRateBps(1, 31)).toEqual({ value: 313, governing_convention_hash: H })
  })
  it.each([
    ['bps_golden_0_15', 0, 15, 0, 'verified'],
    ['bps_golden_16_0', 16, 0, 10000, 'verified'],
    ['bps_golden_1_2', 1, 2, 3333, 'verified'],
    ['bps_golden_16_15', 16, 15, 5161, 'verified'],
    ['bps_golden_0_10', 0, 10, 0, 'verified'],
    ['bps_golden_1_31', 1, 31, 313, 'verified'],
    ['bps_golden_9_23', 9, 23, 2813, 'verified'],
    ['half_even_value_rejected_under_bps', 1, 31, 312, 'rejected'],
    ['old_float_value_under_bps_rejected', 19, 1, 0.95, 'rejected'],
  ])('convention-hash-v0 vector %s', (_n, w, l, value, status) => {
    expect(verifyWinRate(value as number, H, w as number, l as number).status).toBe(status)
  })
  it('unknown convention is unverifiable', () => {
    expect(verifyWinRate(9500, '0x' + 'de'.repeat(32), 19, 1)).toEqual({ status: 'unverifiable' })
  })
})
