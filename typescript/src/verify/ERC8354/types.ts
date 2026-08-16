import type { Address, Hex } from 'viem'

export interface Verdict {
  agentId: bigint
  domainId: Hex
  policyRoot: Hex
  actionCommitment: Hex
  executor: Address
  expiry: bigint
  nullifier: Hex
  decision: number
  policyKind: number
}

export interface PolicyDomain {
  registrar: Address
  verifier: Address
  programKey: Hex
  maxRootAge: bigint
  active: boolean
}

export interface ConfidentialPolicyVerdictClientConfig {
  rpcUrl: string
  address: Address
}

export interface PolicyDomainRegistryClientConfig {
  rpcUrl: string
  address: Address
}
