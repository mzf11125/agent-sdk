import { encodeAbiParameters, keccak256, stringToHex } from 'viem'
import type { Hex } from 'viem'

/**
 * Compute the verdictHash for an ERC-8203 SettlementProofRef.
 *
 * ERC-8203 (ConsultEscrow release commitment):
 *   verdictHash = keccak256(abi.encode(
 *       bytes32 jobId,
 *       keccak256(utf8(resultText))
 *   ))
 *
 * This is the commitment ConsultEscrow.release() recomputes on-chain before
 * checking the attestor signature. The result is fully re-derivable from
 * public data — both jobId and resultText are emitted in the Released event.
 *
 * @param jobId - The 32-byte job identifier (0x-prefixed).
 * @param resultText - The delivered result text (UTF-8).
 * @returns The 32-byte verdict/commitment hash (0x-prefixed).
 */
export function computeVerdictHash(jobId: Hex, resultText: string): Hex {
  const resultHash = keccak256(stringToHex(resultText))
  const encoded = encodeAbiParameters(
    [{ type: 'bytes32' }, { type: 'bytes32' }],
    [jobId, resultHash],
  )
  return keccak256(encoded)
}
