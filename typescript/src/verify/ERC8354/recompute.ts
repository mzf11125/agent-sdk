import { encodeAbiParameters, hashTypedData, keccak256 } from 'viem'
import type { Address, Hex } from 'viem'

/**
 * The source-class tag for a consumed ERC-8354 verdict.
 *
 * ERC-8354 defines `mechanism = keccak256("zk-secret-policy")` as the tag a
 * Guard writes into the ERC-8004 Validation Registry attestation, so verdicts
 * of structurally different guarantees are not conflated into one signal.
 */
export const MECHANISM_ZK_SECRET_POLICY: Hex =
  '0xa843829a78c66c29679817606d0c8a9fa26575b6c2ed0f9f97079d7c46577ac6'

/**
 * Compute the canonical action commitment for a policy verdict.
 *
 * ERC-8354 Action commitment:
 *   actionCommitment = keccak256(abi.encode(chainId, domainId, agentId,
 *       target, value, keccak256(callData), actionNonce))
 *
 * The Guard recomputes this from the action it is about to execute and
 * compares it to `Verdict.actionCommitment`. A Guard must never accept an
 * action commitment supplied by the caller.
 *
 * @param params - The policy action fields.
 * @returns The 32-byte keccak256 action commitment.
 */
export function computeActionCommitment(params: {
  chainId: bigint
  domainId: Hex
  agentId: bigint
  target: Address
  value: bigint
  callData: Hex
  actionNonce: bigint
}): Hex {
  const callDataHash = keccak256(params.callData)
  return keccak256(
    encodeAbiParameters(
      [
        { type: 'uint256' },
        { type: 'bytes32' },
        { type: 'uint256' },
        { type: 'address' },
        { type: 'uint256' },
        { type: 'bytes32' },
        { type: 'uint256' },
      ],
      [
        params.chainId,
        params.domainId,
        params.agentId,
        params.target,
        params.value,
        callDataHash,
        params.actionNonce,
      ],
    ),
  )
}

const verdictTypes = {
  Verdict: [
    { name: 'agentId', type: 'uint256' },
    { name: 'domainId', type: 'bytes32' },
    { name: 'policyRoot', type: 'bytes32' },
    { name: 'actionCommitment', type: 'bytes32' },
    { name: 'executor', type: 'address' },
    { name: 'expiry', type: 'uint64' },
    { name: 'nullifier', type: 'bytes32' },
    { name: 'decision', type: 'uint8' },
    { name: 'policyKind', type: 'uint8' },
  ],
} as const

/**
 * Compute the EIP-712 digest an executor signs to authorize a relayer.
 *
 * ERC-8354 verdictDigest: the EIP-712 typed data hash over the Verdict type
 * with domain name "ConfidentialPolicyVerdict" and version "1". The verdict's
 * single-use nullifier supplies replay protection, so no separate signature
 * nonce is needed.
 *
 * @param verdict - The verdict fields.
 * @param domain - The EIP-712 domain: chainId and the verifying guard contract.
 * @returns The 32-byte EIP-712 verdict digest.
 */
export function computeVerdictDigest(
  verdict: {
    agentId: bigint
    domainId: Hex
    policyRoot: Hex
    actionCommitment: Hex
    executor: Address
    expiry: bigint
    nullifier: Hex
    decision: number
    policyKind: number
  },
  domain: { chainId: number; verifyingContract: Address },
): Hex {
  return hashTypedData({
    domain: {
      name: 'ConfidentialPolicyVerdict',
      version: '1',
      chainId: domain.chainId,
      verifyingContract: domain.verifyingContract,
    },
    types: verdictTypes,
    primaryType: 'Verdict',
    message: verdict,
  })
}
