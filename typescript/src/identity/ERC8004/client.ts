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
