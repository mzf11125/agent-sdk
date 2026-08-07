import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import { computeRawProposalHash, computeVerdictHash } from '../../../src/verify/ERC8299/recomputeL4.js'

// Cross-lane conformance: every language port must reproduce these byte-for-byte.
// Vectors live at testkit/vectors/erc8299-l4.vectors.json (repo-relative, no absolute paths).
const here = dirname(fileURLToPath(import.meta.url))
const VECTORS = resolve(here, '../../../../testkit/vectors/erc8299-l4.vectors.json')

type Vector = {
  step: string
  why: string
  inputs: Record<string, unknown>
  expected: string
}

const suite = JSON.parse(readFileSync(VECTORS, 'utf8')) as { vectors: Vector[] }

describe('ERC-8299 L4 — cross-lane vectors', () => {
  for (const [i, v] of suite.vectors.entries()) {
    it(`[${i}] ${v.step} — ${v.why.slice(0, 70)}`, () => {
      if (v.step === '8299-l4/raw-proposal-hash') {
        expect(computeRawProposalHash(v.inputs.artifact as string)).toBe(v.expected)
      } else if (v.step === '8299-l4/verdict-hash') {
        expect(
          computeVerdictHash(
            v.inputs.fields as Record<string, string | null>,
            v.inputs.preimage_fields as readonly string[],
          ),
        ).toBe(v.expected)
      } else {
        throw new Error(`unknown step ${v.step} — a vector exists that no function covers`)
      }
    })
  }
})
