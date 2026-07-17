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
import { agentVerifierAbi } from './abi.js'
import type { ClientConfig } from './types.js'

export interface VerificationResult {
  valid: boolean
  verificationDigest: Hex
}

export class AgentVerifierClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: ClientConfig['address']
  private readonly abi = agentVerifierAbi

  constructor(config: ClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  // Broadcast, not simulated: unlike ProofVerifierClient.verify(), this
  // records a VerificationCompleted event on-chain — that log is the
  // point of calling it, so a real transaction is warranted here.
  async verify(taskId: Hex, agentId: Hex, inputHash: Hex, outputHash: Hex, proof: Hex): Promise<VerificationResult> {
    const receipt = await this.send('verify', [taskId, agentId, inputHash, outputHash, proof])
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'VerificationCompleted' })
    if (decoded.length === 0) {
      throw new Error('verify: VerificationCompleted event not found in transaction receipt')
    }
    return {
      valid: decoded[0].args.valid as boolean,
      verificationDigest: decoded[0].args.verificationDigest as Hex,
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
