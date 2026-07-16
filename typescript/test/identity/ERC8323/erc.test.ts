import { beforeAll, describe, expect, it } from 'vitest'
import { createPublicClient, createWalletClient, http, parseAbi } from 'viem'
import { privateKeyToAccount } from 'viem/accounts'
import { foundry } from 'viem/chains'
import { SourceBindingViewClient, SOURCE_BINDING_VIEW_INTERFACE_ID } from '../../../src/identity/ERC8323/client.js'
import { IdentityRegistryClient } from '../../../src/identity/ERC8004/client.js'
import { ANVIL_RPC_URL, deployContracts, getAnvilAccount } from '../../setup/deploy.js'

// registerWithSource / mint are not part of SourceBindingViewClient (deliberately
// read-only per spec — source binding is an on-chain fact, no Layer-2 recompute).
// Setup here writes directly against the deployed testkit fixtures, mirroring what
// a real agent-minting flow does before a verifier ever touches the view client.
const setupAbi = parseAbi([
  'function mint(address to) returns (uint256 tokenId)',
  'function registerWithSource(uint256 sourceTokenId) payable returns (uint256 agentId)',
  'function transferFrom(address from, address to, uint256 tokenId)',
])

describe('SourceBindingViewClient (ERC-8323)', () => {
  let client: SourceBindingViewClient
  let sourceAddress: `0x${string}`
  let registryAddress: `0x${string}`

  const holder = privateKeyToAccount(getAnvilAccount(1).privateKey)
  const stranger = privateKeyToAccount(getAnvilAccount(2).privateKey)
  const walletClient = createWalletClient({ chain: foundry, transport: http(ANVIL_RPC_URL) })
  const publicClient = createPublicClient({ chain: foundry, transport: http(ANVIL_RPC_URL) })

  beforeAll(() => {
    ;[sourceAddress, registryAddress] = deployContracts('identity/ERC8323', 'DeployERC8323')
    client = new SourceBindingViewClient({ rpcUrl: ANVIL_RPC_URL, address: registryAddress })
  })

  it('reflects a live registerWithSource mint via the view-only client', async () => {
    const mintHash = await walletClient.writeContract({
      account: holder,
      address: sourceAddress,
      abi: setupAbi,
      functionName: 'mint',
      args: [holder.address],
    })
    await waitForReceipt(mintHash)

    const sourceTokenId = 1n // SourceCollectionMock's first mint

    const registerHash = await walletClient.writeContract({
      account: holder,
      address: registryAddress,
      abi: setupAbi,
      functionName: 'registerWithSource',
      args: [sourceTokenId],
    })
    await waitForReceipt(registerHash)

    const agentId = 1n // MockSourceBindingRegistry's first registration

    expect(await client.hasSourceNFT(agentId)).toBe(true)
    const source = await client.getSourceNFT(agentId)
    expect(source.sourceContract.toLowerCase()).toBe(sourceAddress.toLowerCase())
    expect(source.sourceTokenId).toBe(sourceTokenId)
    expect(await client.isSourceNFTOwnershipValid(agentId)).toBe(true)
    expect(await client.supportsSourceBindingView()).toBe(true)
  })

  it('flips isSourceNFTOwnershipValid to false when the source token is resold, while provenance stays recorded', async () => {
    const mintHash = await walletClient.writeContract({
      account: stranger,
      address: sourceAddress,
      abi: setupAbi,
      functionName: 'mint',
      args: [stranger.address],
    })
    await waitForReceipt(mintHash)
    const sourceTokenId = 2n // second mint on this collection

    const registerHash = await walletClient.writeContract({
      account: stranger,
      address: registryAddress,
      abi: setupAbi,
      functionName: 'registerWithSource',
      args: [sourceTokenId],
    })
    await waitForReceipt(registerHash)
    const agentId = 2n

    expect(await client.isSourceNFTOwnershipValid(agentId)).toBe(true)

    const transferHash = await walletClient.writeContract({
      account: stranger,
      address: sourceAddress,
      abi: setupAbi,
      functionName: 'transferFrom',
      args: [stranger.address, holder.address, sourceTokenId],
    })
    await waitForReceipt(transferHash)

    expect(await client.hasSourceNFT(agentId)).toBe(true) // provenance unchanged
    expect(await client.isSourceNFTOwnershipValid(agentId)).toBe(false) // live check flips
  })

  it('SOURCE_BINDING_VIEW_INTERFACE_ID matches the independently-recomputed spec id', () => {
    expect(SOURCE_BINDING_VIEW_INTERFACE_ID).toBe('0x8b3597c9')
  })

  it('IdentityRegistryClient.registerWithSource mints and is readable via the view client', async () => {
    const mintHash = await walletClient.writeContract({
      account: holder,
      address: sourceAddress,
      abi: setupAbi,
      functionName: 'mint',
      args: [holder.address],
    })
    await waitForReceipt(mintHash)
    const sourceTokenId = 3n // third mint on this collection across this describe block

    const identityClient = new IdentityRegistryClient({ rpcUrl: ANVIL_RPC_URL, address: registryAddress }, holder)
    const agentId = await identityClient.registerWithSource(sourceTokenId)

    expect(await client.hasSourceNFT(agentId)).toBe(true)
    const source = await client.getSourceNFT(agentId)
    expect(source.sourceTokenId).toBe(sourceTokenId)
    expect(await identityClient.ownerOf(agentId)).toBe(holder.address)
  })

  async function waitForReceipt(hash: `0x${string}`) {
    await publicClient.waitForTransactionReceipt({ hash })
  }
})
