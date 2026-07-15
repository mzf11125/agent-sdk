import type { Address } from 'viem'

export interface AgentReputationClientConfig {
  rpcUrl: string
  address: Address
}

export interface ReputationData {
  completedOrders: bigint
  disputedOrders: bigint
  totalVolume: bigint
  lastActiveAt: bigint
  score: number
}
