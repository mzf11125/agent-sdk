import { toHex } from 'viem'

/**
 * Compute the agentId for a given registryId.
 *
 * ERC-8004 / ERC-8299 (PR #1810): agentId = bytes32(uint256(registryId)).
 * This is a left-padded zero-extension of the registry id — not a hash.
 *
 * @param registryId - The registry-assigned agent id (positive integer).
 * @returns The 32-byte left-padded hex representation.
 */
export function computeAgentId(registryId: number | bigint): `0x${string}` {
  return toHex(BigInt(registryId), { size: 32 })
}
