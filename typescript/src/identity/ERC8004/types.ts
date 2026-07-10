import type { Address, Hex } from 'viem'

export interface MetadataEntry {
  metadataKey: string
  metadataValue: Hex
}

export interface IdentityRegistryClientConfig {
  rpcUrl: string
  address: Address
}
