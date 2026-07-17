import {
  createPublicClient,
  createWalletClient,
  http,
  type Account,
  type Address,
  type Hex,
  type PublicClient,
  type WalletClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { agentReputationAbi } from './abi.js'
import type { AgentReputationClientConfig, ReputationData } from './types.js'

export class AgentReputationClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = agentReputationAbi

  constructor(config: AgentReputationClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  /**
   * Read the current reputation snapshot for an agent.
   *
   * @param agentId - The agent's identifier (bytes32).
   * @returns Reputation data: completed/disputed orders, volume, last active timestamp, and score.
   */
  async getReputation(agentId: Hex): Promise<ReputationData> {
    const result = await this.publicClient.readContract({
      address: this.address,
      abi: this.abi,
      functionName: 'getReputation',
      args: [agentId],
    } as never) as unknown as readonly [bigint, bigint, bigint, bigint, number]
    return {
      completedOrders: result[0],
      disputedOrders: result[1],
      totalVolume: result[2],
      lastActiveAt: result[3],
      score: result[4],
    }
  }

  /**
   * Read the recency-decay weight for an agent's score.
   *
   * @param agentId - The agent's identifier (bytes32).
   * @returns Decay weight in basis points (10000 = no decay).
   */
  async getDecayWeight(agentId: Hex): Promise<number> {
    return this.publicClient.readContract({
      address: this.address,
      abi: this.abi,
      functionName: 'getDecayWeight',
      args: [agentId],
    } as never) as Promise<number>
  }

  /**
   * Verify a settled order's outcome proof against the public record.
   *
   * This is a read-only simulated call, not a broadcast transaction: it's a
   * pure cryptographic/on-chain check with no state to persist, so anyone
   * can freely re-derive the answer without spending gas or holding a key
   * with funds.
   *
   * @param orderId - Identifier of the settled order.
   * @param proof - Implementation-defined proof of the settled outcome.
   * @returns True if the outcome is valid against public on-chain data.
   */
  async verifyOutcome(orderId: Hex, proof: Hex): Promise<boolean> {
    const { result } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName: 'verifyOutcome',
      args: [orderId, proof],
      account: this.walletClient.account,
    } as never)
    return result as boolean
  }
}
