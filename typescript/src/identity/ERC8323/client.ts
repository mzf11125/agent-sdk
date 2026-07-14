import {
  createPublicClient,
  http,
  type Address,
  type Hex,
  type PublicClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { agentSourceBindingViewAbi } from './abi.js'
import type { SourceBindingViewClientConfig, SourceNFT } from './types.js'

/**
 * ERC-165 id of `IAgentSourceBindingView`
 * = `getSourceNFT ^ hasSourceNFT ^ isSourceNFTOwnershipValid`.
 * This is the query-only subset a self-sourced ("genesis") agent honestly
 * advertises — NOT the full `IAgentSourceBinding` (`0x27eba962`), which also
 * carries the bridge methods (`boundCollection` / `registerWithSource`).
 */
export const SOURCE_BINDING_VIEW_INTERFACE_ID: Hex = '0x8b3597c9'

/**
 * ERC-8323 Source-Token Agent Binding — read side.
 *
 * View-only client over `IAgentSourceBindingView`. There is no Layer-2 recompute
 * function: source binding is an on-chain fact (the registry maps agentId → the
 * source NFT and validates current ownership), so verification is a direct read,
 * not a hash re-derivation.
 */
export class SourceBindingViewClient {
  private readonly publicClient: PublicClient
  private readonly address: Address
  private readonly abi = agentSourceBindingViewAbi

  constructor(config: SourceBindingViewClientConfig) {
    this.publicClient = createPublicClient({ chain: foundry, transport: http(config.rpcUrl) })
    this.address = config.address
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

  /** ERC-165: does this contract advertise `IAgentSourceBindingView` (`0x8b3597c9`)? */
  async supportsSourceBindingView(): Promise<boolean> {
    return this.read<boolean>('supportsInterface', [SOURCE_BINDING_VIEW_INTERFACE_ID])
  }

  private async read<T>(functionName: string, args: unknown[]): Promise<T> {
    return this.publicClient.readContract({
      address: this.address,
      abi: this.abi,
      functionName,
      args,
    } as never) as Promise<T>
  }
}
