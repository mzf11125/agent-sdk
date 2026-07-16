import {
  createPublicClient,
  createWalletClient,
  http,
  parseEventLogs,
  type Account,
  type Address,
  type Hex,
  type PublicClient,
  type TransactionReceipt,
  type WalletClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { observationCommitmentAbi } from './abi.js'
import type { ObservationCommitmentClientConfig, RecordedEvent } from './types.js'

/**
 * ERC-165 interface id of `IObservationCommitment` (`0xb5c645bd`).
 */
export const OBSERVATION_COMMITMENT_INTERFACE_ID: Hex = '0xb5c645bd'

/**
 * ERC-8281 Observation Commitment Protocol (OCP) — write side.
 *
 * Anchors an opaque commitment digest on-chain via `record()`, emitting a
 * `Recorded` event as tamper-evident proof. The observation itself stays
 * off-chain; verification is recompute-based — a verifier re-derives the
 * digest from the primary artifact and confirms the matching `Recorded` log
 * exists at the claimed chain/block position.
 *
 * There is no on-chain getter: the event log IS the ledger.
 */
export class ObservationCommitmentClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = observationCommitmentAbi

  constructor(config: ObservationCommitmentClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  /**
   * Commit a digest on-chain. Emits `Recorded(digest, committer)`.
   * Returns the transaction receipt so the caller can extract chain/block/log
   * position for the proof envelope.
   */
  async record(digest: Hex): Promise<TransactionReceipt> {
    return this.send('record', [digest])
  }

  /**
   * Convenience: extract the `Recorded` event from a transaction receipt.
   * Throws if no matching event is found.
   */
  parseRecordedEvent(receipt: TransactionReceipt): RecordedEvent {
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'Recorded' })
    if (decoded.length === 0) {
      throw new Error('Recorded event not found in transaction receipt')
    }
    return {
      digest: decoded[0].args.digest as Hex,
      committer: decoded[0].args.committer as Address,
    }
  }

  /** ERC-165: does this contract advertise `IObservationCommitment` (`0xb5c645bd`)? */
  async supportsObservationCommitment(): Promise<boolean> {
    return this.read<boolean>('supportsInterface', [OBSERVATION_COMMITMENT_INTERFACE_ID])
  }

  private async read<T>(functionName: string, args: unknown[]): Promise<T> {
    return this.publicClient.readContract({
      address: this.address,
      abi: this.abi,
      functionName,
      args,
    } as never) as Promise<T>
  }

  private async send(functionName: string, args: unknown[]): Promise<TransactionReceipt> {
    const { request } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName,
      args,
      account: this.walletClient.account,
    } as never)
    const hash = await this.walletClient.writeContract(request as never)
    return this.publicClient.waitForTransactionReceipt({ hash })
  }
}
