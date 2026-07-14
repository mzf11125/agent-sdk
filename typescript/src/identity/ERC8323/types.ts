import type { Address } from 'viem'

export interface SourceBindingViewClientConfig {
  rpcUrl: string
  address: Address
}

export interface SourceNFT {
  sourceContract: Address
  sourceTokenId: bigint
}
