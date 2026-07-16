import { keccak256 } from 'viem'
import type { Hex } from 'viem'

/**
 * Compute `digest = keccak256(observation)`.
 *
 * This is the core OCP commitment step (ERC-8281 §1): the observation bytes
 * are hashed to produce the opaque digest that is anchored on-chain via
 * `record(digest)`.
 *
 * @param observation - The raw observation bytes to commit (Hex).
 * @returns The 32-byte keccak256 digest.
 */
export function computeObservationDigest(observation: Hex): Hex {
  return keccak256(observation)
}
