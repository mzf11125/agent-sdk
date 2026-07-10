import { describe, expect, it } from 'vitest'
import { VERSION } from '../src/version.js'

describe('VERSION', () => {
  it('is a non-empty semver-like string', () => {
    expect(VERSION).toMatch(/^\d+\.\d+\.\d+$/)
  })
})
