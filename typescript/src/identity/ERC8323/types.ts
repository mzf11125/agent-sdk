import type { Address } from 'viem'

export interface SourceBindingClientConfig {
  rpcUrl: string
  address: Address
}

export interface SourceNFT {
  sourceContract: Address
  sourceTokenId: bigint
}
