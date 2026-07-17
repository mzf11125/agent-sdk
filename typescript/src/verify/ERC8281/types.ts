import type { Address } from 'viem'

export interface ObservationCommitmentClientConfig {
  rpcUrl: string
  address: Address
}

export interface RecordedEvent {
  digest: `0x${string}`
  committer: Address
}
