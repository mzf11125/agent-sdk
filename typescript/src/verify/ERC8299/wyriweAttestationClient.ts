import {
  createPublicClient,
  http,
  type Account,
  type Hex,
  type PublicClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { wyriweAttestationAbi } from './abi.js'
import type { ClientConfig, WyriweAttestation } from './types.js'

export class WyriweAttestationClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: { account: Account }
  private readonly address: ClientConfig['address']
  private readonly abi = wyriweAttestationAbi

  constructor(config: ClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = { account }
    this.address = config.address
  }

  // A read-only simulated call, not a broadcast transaction: this checks
  // an EIP-712 signature against the known attestor. Anyone can re-derive
  // the answer without spending gas or holding a funded key — that's the
  // whole point of exposing it this way.
  async verify(attestation: WyriweAttestation, signature: Hex): Promise<boolean> {
    const { result } = await this.publicClient.simulateContract({
      address: this.address,
      abi: this.abi,
      functionName: 'verify',
      args: [
        [
          attestation.agentId,
          attestation.registry,
          attestation.modelHash,
          attestation.rawInputHash,
          attestation.sanitizationPipelineHash,
          attestation.inputHash,
          attestation.outputHash,
          attestation.timestamp,
        ],
        signature,
      ],
      account: this.walletClient.account,
    } as never)
    return result as boolean
  }

  async proofSystem(): Promise<string> {
    return this.publicClient.readContract({
      address: this.address,
      abi: this.abi,
      functionName: 'proofSystem',
    } as never) as Promise<string>
  }
}
