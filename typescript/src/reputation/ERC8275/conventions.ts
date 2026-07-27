/**
 * governing_convention_hash — pin-at-issuance, resolve-at-verification for ERC-8275 winRateBps.
 *
 * A computation rule (formula + representation + rounding mode) is content-addressed:
 *
 *     governing_convention_hash = "0x" + sha256(JCS(convention_spec))
 *
 * because number formatting and rounding are exactly where two honest implementations diverge
 * without disagreeing on the inputs (1 win / 31 losses = 312.5 bps -> 313 half-up vs 312 half-even).
 * A producer pins the hash at issuance; a verifier resolves it and recomputes under *that*
 * convention. Tri-state, fail-closed: an unknown hash is `unverifiable`, never a silent pass.
 *
 * Binds the convention this SDK produces — `win_rate.bps.v0`. The spec object and derived hash are
 * byte-identical to trustless-ai/recompute-kit conformance/convention-hash-v0; the hash is DERIVED
 * from the spec here (reproduce, don't trust) and self-checked against the locked identity on load.
 */
import { sha256, stringToBytes } from 'viem'
import type { Hex } from 'viem'
import { computeWinRate } from './recompute.js'

/** Byte-identical to recompute-kit convention-hash-v0. The hash is derived from THIS object. */
export const WIN_RATE_BPS_V0_SPEC: Record<string, string> = {
  id: 'win_rate.bps.v0',
  quantity: 'erc8275.win_rate',
  formula: 'winRateBps = (gated_wins*20000 + total) // (2*total), total = wins+losses',
  representation: 'integer basis points, 0..10000',
  rounding_mode:
    'round-half-up (half-away-from-zero), exact integer division — never a float round()',
  erc: 'ERC-8275',
  source:
    'agent-sdk#5 @87b08f3 reputation/erc8275 — Python/Rust/TS identical; winRateBps live on babyblueviper /ledger',
}

/** RFC-8785 JCS for a flat string map: sorted keys, compact separators, raw UTF-8. */
function canon(spec: Record<string, string>): string {
  return JSON.stringify(spec, Object.keys(spec).sort())
}

/** Content-address a convention spec: `"0x" + sha256(JCS(spec))`. */
export function governingConventionHash(spec: Record<string, string>): Hex {
  return sha256(stringToBytes(canon(spec)))
}

/** Locked identity (recompute-kit convention-hash-v0) — reproduced from the spec, not trusted. */
export const WIN_RATE_BPS_V0_HASH =
  '0x0501b75db8e9ef4ef67c74efcfbe2a200b0a7e5aea5ca62f778c91c119e68daf' as const

if (governingConventionHash(WIN_RATE_BPS_V0_SPEC).toLowerCase() !== WIN_RATE_BPS_V0_HASH) {
  throw new Error(
    'win_rate.bps.v0 convention-hash drift: derived hash does not match the locked identity',
  )
}

export interface PinnedWinRate {
  value: number
  governing_convention_hash: Hex
}

/** Pin-at-issuance: compute winRateBps and stamp the convention hash that produced it. */
export function pinWinRateBps(wins: number, losses: number): PinnedWinRate {
  return { value: computeWinRate(wins, losses), governing_convention_hash: WIN_RATE_BPS_V0_HASH }
}

export type ConventionStatus = 'verified' | 'rejected' | 'unverifiable'
export interface Verdict {
  status: ConventionStatus
  convention?: string
}

/**
 * Resolve-at-verification, tri-state, fail-closed.
 * verified — value equals the recompute under the pinned convention;
 * rejected — convention resolves but the value disagrees;
 * unverifiable — the hash is not one this SDK produces (never a silent pass).
 */
export function verifyWinRate(
  value: number,
  conventionHash: string,
  wins: number,
  losses: number,
): Verdict {
  if (conventionHash.toLowerCase() !== WIN_RATE_BPS_V0_HASH) {
    return { status: 'unverifiable' }
  }
  const recomputed = computeWinRate(wins, losses)
  return { status: recomputed === value ? 'verified' : 'rejected', convention: 'win_rate.bps.v0' }
}
