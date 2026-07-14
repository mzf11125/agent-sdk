import type { Address, Hex } from 'viem'

export interface ConsultEscrowClientConfig {
  rpcUrl: string
  address: Address
}

export interface Job {
  consumer: Address
  provider: Address
  attestor: Address
  amount: bigint
  deadline: bigint
  status: number // 0=None, 1=Open, 2=Released, 3=Refunded
}

export type JobStatus = 'None' | 'Open' | 'Released' | 'Refunded'

export const JOB_STATUS = {
  None: 0,
  Open: 1,
  Released: 2,
  Refunded: 3,
} as const

export interface OpenedEvent {
  jobId: Hex
  consumer: Address
  provider: Address
  attestor: Address
  amount: bigint
  deadline: bigint
}

export interface ReleasedEvent {
  jobId: Hex
  resultHash: Hex
  commitmentHash: Hex
  provider: Address
  amount: bigint
}

export interface RefundedEvent {
  jobId: Hex
  consumer: Address
  amount: bigint
}
