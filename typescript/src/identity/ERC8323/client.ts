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
import { agentSourceBindingAbi } from './abi.js'
import type { SourceBindingClientConfig, SourceNFT } from './types.js'

/**
 * ERC-165 interface id of `IAgentSourceBinding` (`0x27eba962`).
 * XOR of the five function selectors:
 *   boundCollection ^ getSourceNFT ^ hasSourceNFT
 *   ^ isSourceNFTOwnershipValid ^ registerWithSource.
 */
export const SOURCE_BINDING_INTERFACE_ID: Hex = '0x27eba962'

/**
 * ERC-8323 Source-Token Agent Binding — read + write side.
 *
 * Client over `IAgentSourceBinding`. There is no Layer-2 recompute function:
 * source binding is an on-chain fact (the registry maps agentId -> the source
 * NFT and validates current ownership), so verification is a direct read, not
 * a hash re-derivation.
 */
export class SourceBindingClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = agentSourceBindingAbi

  constructor(config: SourceBindingClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  /** The source ERC-721 collection this registry is bound to. */
  async boundCollection(): Promise<Address> {
    return this.read<Address>('boundCollection', [])
  }

  /** Register an agent derived from `sourceTokenId` of the bound collection. */
  async register(sourceTokenId: bigint): Promise<bigint> {
    const receipt = await this.send('registerWithSource', [sourceTokenId])
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'SourceNFTLinked' })
    if (decoded.length === 0) {
      throw new Error('register: SourceNFTLinked event not found in transaction receipt')
    }
    return decoded[0].args.agentId as bigint
  }

  /** The source NFT `(contract, tokenId)` an agent is bound to. */
  async getSourceNFT(agentId: bigint): Promise<SourceNFT> {
    const [sourceContract, sourceTokenId] = await this.read<[Address, bigint]>('getSourceNFT', [agentId])
    return { sourceContract, sourceTokenId }
  }

  /** Whether the agent claims a source NFT (false for unbound / self-sourced-only). */
  async hasSourceNFT(agentId: bigint): Promise<boolean> {
    return this.read<boolean>('hasSourceNFT', [agentId])
  }

  /** Whether the bound source NFT is still owned by the agent's controller. */
  async isSourceNFTOwnershipValid(agentId: bigint): Promise<boolean> {
    return this.read<boolean>('isSourceNFTOwnershipValid', [agentId])
  }

  /** ERC-165: does this contract advertise `IAgentSourceBinding` (`0x27eba962`)? */
  async supportsSourceBinding(): Promise<boolean> {
    return this.read<boolean>('supportsInterface', [SOURCE_BINDING_INTERFACE_ID])
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
