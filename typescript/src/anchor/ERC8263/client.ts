import {
  createPublicClient,
  createWalletClient,
  http,
  type Account,
  type Address,
  type Hex,
  type PublicClient,
  type TransactionReceipt,
  type WalletClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { onChainProofAbi } from './abi.js'
import type { OnChainProofClientConfig } from './types.js'

export class OnChainProofClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = onChainProofAbi

  constructor(config: OnChainProofClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  async anchor(agentIdScheme: number, agentId: Hex, proofHash: Hex): Promise<TransactionReceipt> {
    return this.send('anchor', [agentIdScheme, agentId, proofHash])
  }

  async anchorWithAux(
    agentIdScheme: number,
    agentId: Hex,
    proofHash: Hex,
    aux: Hex,
  ): Promise<TransactionReceipt> {
    return this.send('anchorWithAux', [agentIdScheme, agentId, proofHash, aux])
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
