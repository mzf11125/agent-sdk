import type { Address, Hex } from 'viem'

export interface ClientConfig {
  rpcUrl: string
  address: Address
}

export interface WyriweAttestation {
  agentId: Hex
  registry: Address
  modelHash: Hex
  rawInputHash: Hex
  sanitizationPipelineHash: Hex
  inputHash: Hex
  outputHash: Hex
  timestamp: bigint
}

export interface JudgmentExecutionAttestation {
  agentId: Hex
  registry: Address
  validatorId: Hex
  rawProposalHash: Hex
  verdictHash: Hex
  executedActionHash: Hex
  verdictTimestamp: bigint
  executedTimestamp: bigint
  recordPointer: string
}
