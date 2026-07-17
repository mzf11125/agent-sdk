import {
  createPublicClient,
  createWalletClient,
  http,
  type Account,
  type Hex,
  type PublicClient,
  type WalletClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { proofVerifierAbi } from './abi.js'
import type { ClientConfig } from './types.js'

export class ProofVerifierClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: ClientConfig['address']
  private readonly abi = proofVerifierAbi

  constructor(config: ClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  // A read-only simulated call, not a broadcast transaction: this is a
  // pure cryptographic check with no state to persist, so anyone can
  // freely re-derive the answer without spending gas or holding a key
  // with funds — that's the whole point of exposing it this way.
  async verify(inputHash: Hex, outputHash: Hex, metadata: Hex, proof: Hex): Promise<boolean> {
    const { result } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName: 'verify',
      args: [inputHash, outputHash, metadata, proof],
      account: this.walletClient.account,
    } as never)
    return result as boolean
  }

  async proofSystem(): Promise<string> {
    return this.read('proofSystem', [])
  }

  async proofProfile(): Promise<Hex> {
    return this.read('proofProfile', [])
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
