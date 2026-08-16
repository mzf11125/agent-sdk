import { beforeAll, describe, expect, it } from 'vitest'
import {
  createPublicClient,
  createWalletClient,
  http,
  keccak256,
  stringToHex,
  type Address,
  type Hex,
} from 'viem'
import { foundry } from 'viem/chains'
import { privateKeyToAccount } from 'viem/accounts'
import {
  ConfidentialPolicyVerdictClient,
  PolicyDomainRegistryClient,
} from '../../../src/verify/ERC8354/client.js'
import { computeActionCommitment, computeVerdictDigest } from '../../../src/verify/ERC8354/recompute.js'
import type { Verdict } from '../../../src/verify/ERC8354/types.js'
import {
  ANVIL_RPC_URL,
  deployContracts,
  getAnvilAccount,
} from '../../setup/deploy.js'

const DOMAIN_ID = keccak256(stringToHex('test-domain'))
const POLICY_ROOT = keccak256(stringToHex('root-v1'))
const PROGRAM_KEY = keccak256(stringToHex('interpreter-vkey'))

// The concrete registry's admin methods, used only for test setup. They are
// not part of IPolicyDomainRegistry, so they are not exposed by the SDK.
const registryAdminAbi = [
  {
    type: 'function',
    name: 'registerDomain',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'domainId', type: 'bytes32' },
      { name: 'registrar', type: 'address' },
      { name: 'verifier', type: 'address' },
      { name: 'programKey', type: 'bytes32' },
      { name: 'maxRootAge', type: 'uint64' },
    ],
    outputs: [],
  },
  {
    type: 'function',
    name: 'updateRoot',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'domainId', type: 'bytes32' },
      { name: 'newRoot', type: 'bytes32' },
    ],
    outputs: [],
  },
] as const

// The mock verifier's setResult toggle, used only for the malformed/invalid
// proof negative path. It is not part of any ERC interface, so it is not
// exposed by the SDK.
const mockVerifierAbi = [
  {
    type: 'function',
    name: 'setResult',
    stateMutability: 'nonpayable',
    inputs: [{ name: 'r', type: 'bool' }],
    outputs: [],
  },
] as const

describe('ConfidentialPolicyVerdictClient (ERC-8354)', () => {
  let guardClient: ConfidentialPolicyVerdictClient
  let registryClient: PolicyDomainRegistryClient
  let verifierAddress: Address
  let guardAddress: Address

  const admin = privateKeyToAccount(getAnvilAccount(0).privateKey)
  const executor = privateKeyToAccount(getAnvilAccount(1).privateKey)

  const actionCommitment = computeActionCommitment({
    chainId: 31337n,
    domainId: DOMAIN_ID,
    agentId: 1n,
    target: '0x0000000000000000000000000000000000000001',
    value: 0n,
    callData: '0x',
    actionNonce: 0n,
  })

  function verdict(nullifierSeed: string): Verdict {
    return {
      agentId: 1n,
      domainId: DOMAIN_ID,
      policyRoot: POLICY_ROOT,
      actionCommitment,
      executor: executor.address,
      expiry: 4_000_000_000n,
      nullifier: keccak256(stringToHex(nullifierSeed)),
      decision: 1,
      policyKind: 0,
    }
  }

  beforeAll(async () => {
    const [verifier, registry, guard] = deployContracts('verify/ERC8354', 'DeployERC8354')
    verifierAddress = verifier
    guardAddress = guard

    const publicClient = createPublicClient({ chain: foundry, transport: http(ANVIL_RPC_URL) })
    const walletClient = createWalletClient({
      chain: foundry,
      transport: http(ANVIL_RPC_URL),
      account: admin,
    })

    const registerRequest = await publicClient.simulateContract({
      address: registry,
      abi: registryAdminAbi,
      functionName: 'registerDomain',
      args: [DOMAIN_ID, '0x000000000000000000000000000000000000a11c', verifierAddress, PROGRAM_KEY, 3600n],
      account: admin,
    } as never)
    const registerHash = await walletClient.writeContract(registerRequest.request as never)
    // Wait for the domain to exist before issuing the root update, which
    // depends on it. Broadcasting both without waiting is a timing race.
    await publicClient.waitForTransactionReceipt({ hash: registerHash })

    const updateRequest = await publicClient.simulateContract({
      address: registry,
      abi: registryAdminAbi,
      functionName: 'updateRoot',
      args: [DOMAIN_ID, POLICY_ROOT],
      account: admin,
    } as never)
    const updateHash = await walletClient.writeContract(updateRequest.request as never)
    await publicClient.waitForTransactionReceipt({ hash: updateHash })

    guardClient = new ConfidentialPolicyVerdictClient(
      { rpcUrl: ANVIL_RPC_URL, address: guardAddress },
      executor,
    )
    registryClient = new PolicyDomainRegistryClient({ rpcUrl: ANVIL_RPC_URL, address: registry })
  })

  it('supports the confidential policy verdict interface', async () => {
    expect(await guardClient.supportsConfidentialPolicyVerdict()).toBe(true)
  })

  it('reports the domain root as acceptable', async () => {
    expect(await registryClient.isRootAcceptable(DOMAIN_ID, POLICY_ROOT)).toBe(true)
  })

  it('reads the registered domain', async () => {
    const domain = await registryClient.domain(DOMAIN_ID)
    expect(domain.active).toBe(true)
    expect(domain.verifier.toLowerCase()).toBe(verifierAddress.toLowerCase())
  })

  it('verifies a valid verdict', async () => {
    expect(await guardClient.verify(verdict('nf-verify'), '0x1234')).toBe(true)
  })

  it('recomputes the verdict digest', async () => {
    const digest = await guardClient.verdictDigest(verdict('nf-digest'))
    expect(digest).toMatch(/^0x[0-9a-f]{64}$/)
  })

  it('consumes a verdict and burns the nullifier', async () => {
    const v = verdict('nf-consume')
    const receipt = await guardClient.consume(v, '0x1234')
    expect(receipt.status).toBe('success')
    expect(await guardClient.isConsumed(DOMAIN_ID, v.nullifier)).toBe(true)
  })

  it('rejects a replayed verdict', async () => {
    const v = verdict('nf-replay')
    const first = await guardClient.consume(v, '0x1234')
    expect(first.status).toBe('success')

    await expect(guardClient.consume(v, '0x1234')).rejects.toThrow()
  })

  it('consumes a verdict via a relayer', async () => {
    // The executor (account 1) authorizes a relayer (account 0) by signing
    // the EIP-712 verdict digest. The relayer submits the transaction, so
    // msg.sender is not the executor and the signature must validate.
    const v = verdict('nf-relayed')
    const digest = computeVerdictDigest(v, { chainId: 31337, verifyingContract: guardAddress })
    const executorAuth = await executor.sign({ hash: digest })

    const relayClient = new ConfidentialPolicyVerdictClient(
      { rpcUrl: ANVIL_RPC_URL, address: guardAddress },
      admin,
    )
    const receipt = await relayClient.consumeRelayed(v, '0x1234', executorAuth)
    expect(receipt.status).toBe('success')
    expect(await relayClient.isConsumed(DOMAIN_ID, v.nullifier)).toBe(true)
  })

  it('rejects a relayer with a bad executor signature', async () => {
    const v = verdict('nf-bad-sig')
    const relayClient = new ConfidentialPolicyVerdictClient(
      { rpcUrl: ANVIL_RPC_URL, address: guardAddress },
      admin,
    )
    await expect(relayClient.consumeRelayed(v, '0x1234', '0x1234')).rejects.toThrow()
  })

  it('rejects an invalid proof', async () => {
    // Force the domain verifier to reject the proof, then a direct consume
    // must revert with InvalidProof rather than burn the nullifier.
    const adminPublic = createPublicClient({ chain: foundry, transport: http(ANVIL_RPC_URL) })
    const adminWallet = createWalletClient({
      chain: foundry,
      transport: http(ANVIL_RPC_URL),
      account: admin,
    })
    const setRequest = await adminPublic.simulateContract({
      address: verifierAddress,
      abi: mockVerifierAbi,
      functionName: 'setResult',
      args: [false],
      account: admin,
    } as never)
    const setHash = await adminWallet.writeContract(setRequest.request as never)
    await adminPublic.waitForTransactionReceipt({ hash: setHash })

    const v = verdict('nf-invalid-proof')
    await expect(guardClient.consume(v, '0x1234')).rejects.toThrow()
  })
})
