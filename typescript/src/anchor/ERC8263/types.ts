import type { Address } from 'viem'

export interface OnChainProofClientConfig {
  rpcUrl: string
  address: Address
}
