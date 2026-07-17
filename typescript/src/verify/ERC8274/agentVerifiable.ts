import { createPublicClient, http, type Address } from 'viem'
import { foundry } from 'viem/chains'
import { agentVerifiableAbi } from './abi.js'

export interface AgentVerifiableConfig {
  rpcUrl: string
  address: Address
}

// A standalone function, not a client class: IAgentVerifiable is a single
// getter, so a class would be one method wrapping a constructor for no
// benefit.
export async function getTrustedVerifier(config: AgentVerifiableConfig): Promise<Address> {
  const publicClient = createPublicClient({ chain: foundry, transport: http(config.rpcUrl) })
  return publicClient.readContract({
    address: config.address,
    abi: agentVerifiableAbi,
    functionName: 'agentVerifier',
  } as never) as Promise<Address>
}
