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
import { consultEscrowAbi } from './abi.js'
import { computeVerdictHash } from './recompute.js'
import type { ConsultEscrowClientConfig, Job, OpenedEvent, ReleasedEvent, RefundedEvent } from './types.js'

export class ConsultEscrowClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = consultEscrowAbi

  constructor(config: ConsultEscrowClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  /**
   * Open a new escrow job. The caller (msg.sender) becomes the consumer.
   *
   * @param jobId - Unique 32-byte job identifier.
   * @param provider - The agent wallet that receives payment on release.
   * @param attestor - The key whose signature authorizes release.
   * @param deadline - Unix timestamp after which the consumer may refund.
   * @param value - Amount of ETH to lock (in wei).
   * @returns The transaction receipt.
   */
  async open(jobId: Hex, provider: Address, attestor: Address, deadline: bigint, value: bigint): Promise<TransactionReceipt> {
    const { request } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName: 'open',
      args: [jobId, provider, attestor, deadline],
      account: this.walletClient.account,
      value,
    } as never)
    const hash = await this.walletClient.writeContract(request as never)
    return this.publicClient.waitForTransactionReceipt({ hash })
  }

  /**
   * Release escrowed funds to the provider, authorized by the attestor's
   * signature over `commitmentHash = keccak256(abi.encode(jobId, resultHash))`.
   *
   * @param jobId - The job to settle.
   * @param resultHash - keccak256 of the delivered result text.
   * @param signature - EIP-191 personal_sign by the attestor over the commitment.
   * @returns The transaction receipt.
   */
  async release(jobId: Hex, resultHash: Hex, signature: Hex): Promise<TransactionReceipt> {
    const { request } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName: 'release',
      args: [jobId, resultHash, signature],
      account: this.walletClient.account,
    } as never)
    const hash = await this.walletClient.writeContract(request as never)
    return this.publicClient.waitForTransactionReceipt({ hash })
  }

  /**
   * Refund the consumer after the deadline has passed.
   *
   * @param jobId - The job to refund.
   * @returns The transaction receipt.
   */
  async refund(jobId: Hex): Promise<TransactionReceipt> {
    const { request } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName: 'refund',
      args: [jobId],
      account: this.walletClient.account,
    } as never)
    const hash = await this.walletClient.writeContract(request as never)
    return this.publicClient.waitForTransactionReceipt({ hash })
  }

  /**
   * Read the escrowed job details.
   *
   * @param jobId - The job identifier.
   * @returns The Job struct with consumer, provider, attestor, amount, deadline, and status.
   */
  async getJob(jobId: Hex): Promise<Job> {
    const result = await this.publicClient.readContract({
      address: this.address,
      abi: this.abi,
      functionName: 'jobs',
      args: [jobId],
    } as never) as unknown as readonly [Address, Address, Address, bigint, bigint, number]
    return {
      consumer: result[0],
      provider: result[1],
      attestor: result[2],
      amount: result[3],
      deadline: result[4],
      status: result[5],
    }
  }

  /**
   * Verify a claimed commitment hash against an independently recomputed one.
   *
   * This is a pure recompute-to-verify check (Layer 2) — it does not call
   * the contract and requires no gas or RPC connection. It recomputes
   * `verdictHash = keccak256(abi.encode(jobId, keccak256(utf8(resultText))))`
   * and compares it to the claimed commitmentHash.
   *
   * @param commitmentHash - The claimed commitment hash (e.g. from a Released event).
   * @param jobId - The job identifier.
   * @param resultText - The delivered result text.
   * @returns True if the recomputed hash matches the claimed one.
   */
  verify(commitmentHash: Hex, jobId: Hex, resultText: string): boolean {
    const recomputed = computeVerdictHash(jobId, resultText)
    return recomputed === commitmentHash
  }

  /**
   * Parse Opened events from a transaction receipt.
   */
  parseOpened(receipt: TransactionReceipt): OpenedEvent[] {
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'Opened' })
    return decoded.map((log) => ({
      jobId: log.args.jobId as Hex,
      consumer: log.args.consumer as Address,
      provider: log.args.provider as Address,
      attestor: log.args.attestor as Address,
      amount: log.args.amount as bigint,
      deadline: log.args.deadline as bigint,
    }))
  }

  /**
   * Parse Released events from a transaction receipt.
   */
  parseReleased(receipt: TransactionReceipt): ReleasedEvent[] {
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'Released' })
    return decoded.map((log) => ({
      jobId: log.args.jobId as Hex,
      resultHash: log.args.resultHash as Hex,
      commitmentHash: log.args.commitmentHash as Hex,
      provider: log.args.provider as Address,
      amount: log.args.amount as bigint,
    }))
  }

  /**
   * Parse Refunded events from a transaction receipt.
   */
  parseRefunded(receipt: TransactionReceipt): RefundedEvent[] {
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'Refunded' })
    return decoded.map((log) => ({
      jobId: log.args.jobId as Hex,
      consumer: log.args.consumer as Address,
      amount: log.args.amount as bigint,
    }))
  }
}
