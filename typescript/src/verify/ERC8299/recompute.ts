import { concat, keccak256, stringToHex } from 'viem'

/**
 * Compute the raw input hash for a WYRIWE attestation.
 *
 * ERC-8299 Section 45: raw_input_hash = keccak256(raw_user_input).
 *
 * @param rawInputHex - The raw user input as a 0x-prefixed hex string.
 * @returns The 32-byte keccak256 hash of the raw input.
 */
export function computeRawInputHash(rawInputHex: `0x${string}`): `0x${string}` {
  return keccak256(rawInputHex)
}

/**
 * Compute the sanitization pipeline hash for a WYRIWE attestation.
 *
 * ERC-8299 Section 46: sanitization_pipeline_hash =
 *     keccak256(utf8(cid) || raw_input_hash).
 *
 * The spec_cid is converted to UTF-8 bytes, then concatenated with the
 * raw_input_hash bytes before hashing.
 *
 * @param specCid - The sanitization spec CID string (e.g. "ipfs://Qm...").
 * @param rawInputHash - The raw input hash (32 bytes, 0x-prefixed).
 * @returns The 32-byte keccak256 hash of the concatenated bytes.
 */
export function computeSanitizationPipelineHash(
  specCid: string,
  rawInputHash: `0x${string}`,
): `0x${string}` {
  const cidBytes = stringToHex(specCid)
  return keccak256(concat([cidBytes, rawInputHash]))
}
