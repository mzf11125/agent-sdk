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
import { contestableEnvelopeAbi } from './abi.js'
import type { ClientConfig } from './types.js'

export interface EnvelopeContestedEvent {
  id: Hex
  challenger: `0x${string}`
}

export interface EnvelopeResolvedEvent {
  id: Hex
  outcome: number
}

export class ContestableEnvelopeClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: ClientConfig['address']
  private readonly abi = contestableEnvelopeAbi

  constructor(config: ClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  async contest(id: Hex, evidence: Hex): Promise<EnvelopeContestedEvent> {
    const receipt = await this.send('contest', [id, evidence])
    const decoded = parseEventLogs({
      abi: this.abi,
      logs: receipt.logs,
      eventName: 'EnvelopeContested',
    })
    if (decoded.length === 0) {
      throw new Error('contest: EnvelopeContested event not found')
    }
    return {
      id: decoded[0].args.id as Hex,
      challenger: decoded[0].args.challenger as `0x${string}`,
    }
  }

  async resolve(id: Hex, outcome: number, resolution: Hex): Promise<EnvelopeResolvedEvent> {
    const receipt = await this.send('resolve', [id, outcome, resolution])
    const decoded = parseEventLogs({
      abi: this.abi,
      logs: receipt.logs,
      eventName: 'EnvelopeResolved',
    })
    if (decoded.length === 0) {
      throw new Error('resolve: EnvelopeResolved event not found')
    }
    return {
      id: decoded[0].args.id as Hex,
      outcome: decoded[0].args.outcome as number,
    }
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
