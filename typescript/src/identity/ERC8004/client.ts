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
import { identityRegistryAbi } from './abi.js'
import type { IdentityRegistryClientConfig, MetadataEntry } from './types.js'

export class IdentityRegistryClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = identityRegistryAbi

  constructor(config: IdentityRegistryClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  async register(agentURI = '', metadata: MetadataEntry[] = []): Promise<bigint> {
    const receipt = await this.send('register', [agentURI, metadata])
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'Registered' })
    if (decoded.length === 0) {
      throw new Error('register: Registered event not found in transaction receipt')
    }
    return decoded[0].args.agentId as bigint
  }

  /**
   * ERC-8323 (Source-Token Agent Binding) — mints an agent bound to
   * `sourceTokenId` on this registry's fixed source collection. This is the
   * single spec-defined overload (`registerWithSource(uint256)`); a
   * registry MAY require payment (the ABI marks it `payable`) — pass
   * `value` if the deployment enforces a mint price, omit it (defaults to
   * 0) for a free registry like this repo's own MockSourceBindingRegistry.
   * Two deployment-specific extension overloads exist below
   * (`registerWithSourceAndURI` / `registerWithSourceAndMetadata`) for
   * registries known to expose them (confirmed on Merlini's
   * AgentIdentityRegistry, 2026-07-16) — NOT part of the ERC-8323 base
   * interface, don't assume a given registry has them.
   */
  async registerWithSource(sourceTokenId: bigint, value: bigint = 0n): Promise<bigint> {
    const receipt = await this.send('registerWithSource', [sourceTokenId], value)
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'SourceNFTLinked' })
    if (decoded.length === 0) {
      throw new Error('registerWithSource: SourceNFTLinked event not found in transaction receipt')
    }
    return decoded[0].args.agentId as bigint
  }

  /**
   * Deployment-specific extension of registerWithSource (NOT ERC-8323 base
   * spec) that lets the caller override the minted agent's URI instead of
   * the registry's baseAgentURI template. Confirmed live on Merlini's
   * AgentIdentityRegistry (mainnet 0xe0454dfa17a57a84c3e0e2dbfda5318cbbe91e2c,
   * 2026-07-16 Telegram) — only call this against a registry known to
   * implement the exact same overload; a bare ERC-8323 registry does not
   * have to expose it.
   */
  async registerWithSourceAndURI(
    agentURI: string,
    sourceTokenId: bigint,
    value: bigint = 0n,
  ): Promise<bigint> {
    const receipt = await this.send('registerWithSource', [agentURI, sourceTokenId], value)
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'SourceNFTLinked' })
    if (decoded.length === 0) {
      throw new Error('registerWithSourceAndURI: SourceNFTLinked event not found in transaction receipt')
    }
    return decoded[0].args.agentId as bigint
  }

  /**
   * Deployment-specific extension of registerWithSource (NOT ERC-8323 base
   * spec) that seeds initial metadata entries at registration time, on top
   * of the custom-URI overload above. Same provenance/caveat as
   * `registerWithSourceAndURI`.
   */
  async registerWithSourceAndMetadata(
    agentURI: string,
    sourceTokenId: bigint,
    metadata: MetadataEntry[],
    value: bigint = 0n,
  ): Promise<bigint> {
    const receipt = await this.send('registerWithSource', [agentURI, sourceTokenId, metadata], value)
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'SourceNFTLinked' })
    if (decoded.length === 0) {
      throw new Error('registerWithSourceAndMetadata: SourceNFTLinked event not found in transaction receipt')
    }
    return decoded[0].args.agentId as bigint
  }

  async setAgentURI(agentId: bigint, agentURI: string): Promise<void> {
    await this.send('setAgentURI', [agentId, agentURI])
  }

  async getAgentURI(agentId: bigint): Promise<string> {
    return this.read('tokenURI', [agentId])
  }

  async getMetadata(agentId: bigint, metadataKey: string): Promise<Hex> {
    return this.read('getMetadata', [agentId, metadataKey])
  }

  async setMetadata(agentId: bigint, metadataKey: string, metadataValue: Hex): Promise<void> {
    await this.send('setMetadata', [agentId, metadataKey, metadataValue])
  }

  async setAgentWallet(agentId: bigint, newWallet: Address, deadline: bigint, signature: Hex): Promise<void> {
    await this.send('setAgentWallet', [agentId, newWallet, deadline, signature])
  }

  async getAgentWallet(agentId: bigint): Promise<Address> {
    return this.read('getAgentWallet', [agentId])
  }

  async unsetAgentWallet(agentId: bigint): Promise<void> {
    await this.send('unsetAgentWallet', [agentId])
  }

  async ownerOf(agentId: bigint): Promise<Address> {
    return this.read('ownerOf', [agentId])
  }

  private async read<T>(functionName: string, args: unknown[]): Promise<T> {
    return this.publicClient.readContract({
      address: this.address,
      abi: this.abi,
      functionName,
      args,
    } as never) as Promise<T>
  }

  private async send(functionName: string, args: unknown[], value: bigint = 0n): Promise<TransactionReceipt> {
    const { request } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName,
      args,
      value,
      account: this.walletClient.account,
    } as never)
    const hash = await this.walletClient.writeContract(request as never)
    return this.publicClient.waitForTransactionReceipt({ hash })
  }
}
