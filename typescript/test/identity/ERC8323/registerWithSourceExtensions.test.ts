import { beforeAll, describe, expect, it } from 'vitest'
import { createPublicClient, createWalletClient, http, parseAbi } from 'viem'
import { privateKeyToAccount } from 'viem/accounts'
import { foundry } from 'viem/chains'
import { IdentityRegistryClient } from '../../../src/identity/ERC8004/client.js'
import { ANVIL_RPC_URL, deployContracts, getAnvilAccount } from '../../setup/deploy.js'

// Confirms the two registerWithSource extension overloads Merlini gave us for his real
// AgentIdentityRegistry (mainnet 0xe0454dfa17a57a84c3e0e2dbfda5318cbbe91e2c, 2026-07-16
// Telegram) actually work end-to-end -- including that payment is genuinely threaded
// through send() now, not just typed. Runs against MockSourceBindingRegistryExtended
// (a separate fixture from the base-spec mock, see that contract's own header comment
// for why it's kept separate).
const MINT_PRICE = 1_000_000_000_000_000n // 0.001 ether, matches DeployERC8323Extended

const setupAbi = parseAbi(['function mint(address to) returns (uint256 tokenId)'])

describe('IdentityRegistryClient registerWithSource extension overloads', () => {
  let sourceAddress: `0x${string}`
  let registryAddress: `0x${string}`

  const holder = privateKeyToAccount(getAnvilAccount(3).privateKey)
  const walletClient = createWalletClient({ chain: foundry, transport: http(ANVIL_RPC_URL) })
  const publicClient = createPublicClient({ chain: foundry, transport: http(ANVIL_RPC_URL) })

  beforeAll(() => {
    ;[sourceAddress, registryAddress] = deployContracts('identity/ERC8323', 'DeployERC8323Extended')
  })

  async function mintSourceToken(): Promise<void> {
    const hash = await walletClient.writeContract({
      account: holder,
      address: sourceAddress,
      abi: setupAbi,
      functionName: 'mint',
      args: [holder.address],
    })
    await publicClient.waitForTransactionReceipt({ hash })
  }

  it('registerWithSource (base overload) reverts with wrong value, succeeds with correct mintPrice', async () => {
    await mintSourceToken()

    const client = new IdentityRegistryClient({ rpcUrl: ANVIL_RPC_URL, address: registryAddress }, holder)

    await expect(client.registerWithSource(1n, 0n)).rejects.toThrow()
    const agentId = await client.registerWithSource(1n, MINT_PRICE)
    expect(agentId).toBe(1n)
  })

  it('registerWithSourceAndURI mints with a custom agentURI, correct value required', async () => {
    await mintSourceToken()

    const client = new IdentityRegistryClient({ rpcUrl: ANVIL_RPC_URL, address: registryAddress }, holder)
    const agentId = await client.registerWithSourceAndURI('ipfs://custom-uri', 2n, MINT_PRICE)

    const extendedAbi = parseAbi(['function agentURI(uint256) view returns (string)'])
    const uri = await publicClient.readContract({
      address: registryAddress,
      abi: extendedAbi,
      functionName: 'agentURI',
      args: [agentId],
    })
    expect(uri).toBe('ipfs://custom-uri')
  })

  it('registerWithSourceAndMetadata mints with a custom URI and seeds metadata', async () => {
    await mintSourceToken()

    const client = new IdentityRegistryClient({ rpcUrl: ANVIL_RPC_URL, address: registryAddress }, holder)
    const agentId = await client.registerWithSourceAndMetadata(
      'ipfs://with-metadata',
      3n,
      [{ metadataKey: 'role', metadataValue: '0x01' }],
      MINT_PRICE,
    )

    const extendedAbi = parseAbi(['function getMetadata(uint256, string) view returns (bytes)'])
    const value = await publicClient.readContract({
      address: registryAddress,
      abi: extendedAbi,
      functionName: 'getMetadata',
      args: [agentId, 'role'],
    })
    expect(value).toBe('0x01')
  })
})
