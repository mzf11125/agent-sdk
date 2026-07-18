import {
  createPublicClient,
  createWalletClient,
  http,
  parseEventLogs,
  type Account,
  type Hex,
  type PublicClient,
  type TransactionReceipt,
  type WalletClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { boundedAgentActionAbi } from './abi.js'
import type { ClientConfig } from './types.js'

export interface Envelope {
  id: Hex
  principal: `0x${string}`
  capabilityRoot: Hex
  cursorRoot: Hex
  createdAt: bigint
  expiresAt: bigint
  status: number // 0=None, 1=Active, 2=Completed, 3=Contested, 4=Revoked, 5=Expired
}

export interface EnvelopeRegisteredEvent {
  id: Hex
  principal: `0x${string}`
  capabilityRoot: Hex
}

export interface EnvelopeAdvancedEvent {
  id: Hex
  prevCursor: Hex
  newCursor: Hex
}

export class BoundedAgentActionClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: ClientConfig['address']
  private readonly abi = boundedAgentActionAbi

  constructor(config: ClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  async registerEnvelope(
    principal: `0x${string}`,
    capabilityRoot: Hex,
    expiresAt: bigint,
    initData: Hex,
  ): Promise<EnvelopeRegisteredEvent> {
    const receipt = await this.send('registerEnvelope', [principal, capabilityRoot, expiresAt, initData])
    const decoded = parseEventLogs({
      abi: this.abi,
      logs: receipt.logs,
      eventName: 'EnvelopeRegistered',
    })
    if (decoded.length === 0) {
      throw new Error('registerEnvelope: EnvelopeRegistered event not found')
    }
    return {
      id: decoded[0].args.id as Hex,
      principal: decoded[0].args.principal as `0x${string}`,
      capabilityRoot: decoded[0].args.capabilityRoot as Hex,
    }
  }

  async advanceCursor(id: Hex, witness: Hex): Promise<EnvelopeAdvancedEvent> {
    const receipt = await this.send('advanceCursor', [id, witness])
    const decoded = parseEventLogs({
      abi: this.abi,
      logs: receipt.logs,
      eventName: 'EnvelopeAdvanced',
    })
    if (decoded.length === 0) {
      throw new Error('advanceCursor: EnvelopeAdvanced event not found')
    }
    return {
      id: decoded[0].args.id as Hex,
      prevCursor: decoded[0].args.prevCursor as Hex,
      newCursor: decoded[0].args.newCursor as Hex,
    }
  }

  async setStatus(id: Hex, newStatus: number): Promise<TransactionReceipt> {
    return this.send('setStatus', [id, newStatus])
  }

  async getEnvelope(id: Hex): Promise<Envelope> {
    const result = await this.read('getEnvelope', [id])
    const env = result as Record<string, unknown>
    return {
      id: env.id as Hex,
      principal: env.principal as `0x${string}`,
      capabilityRoot: env.capabilityRoot as Hex,
      cursorRoot: env.cursorRoot as Hex,
      createdAt: env.createdAt as bigint,
      expiresAt: env.expiresAt as bigint,
      status: env.status as number,
    }
  }

  async getCursor(id: Hex): Promise<Hex> {
    return this.read('getCursor', [id]) as Promise<Hex>
  }

  async getStatus(id: Hex): Promise<number> {
    return this.read('getStatus', [id]) as Promise<number>
  }

  async isActive(id: Hex): Promise<boolean> {
    return this.read('isActive', [id]) as Promise<boolean>
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
