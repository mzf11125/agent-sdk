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
import { confidentialPolicyVerdictAbi, policyDomainRegistryAbi } from './abi.js'
import type {
  ConfidentialPolicyVerdictClientConfig,
  PolicyDomain,
  PolicyDomainRegistryClientConfig,
  Verdict,
} from './types.js'

/**
 * ERC-165 interface id of `IConfidentialPolicyVerdict` (`0xd6da8150`).
 */
export const CONFIDENTIAL_POLICY_VERDICT_INTERFACE_ID: Hex = '0xd6da8150'

/**
 * ERC-8354 Confidential Policy Verdicts, guard side.
 *
 * Consumes a zero-knowledge verdict that an action was evaluated against a
 * committed secret policy and permitted. `verify` is a read-only view call;
 * `consume` burns the verdict's single-use nullifier and gates execution.
 */
export class ConfidentialPolicyVerdictClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = confidentialPolicyVerdictAbi

  constructor(config: ConfidentialPolicyVerdictClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  /**
   * Verify a verdict without state change. Returns false on a well-formed but
   * invalid verdict, and never reverts on a malformed proof.
   */
  async verify(v: Verdict, proof: Hex): Promise<boolean> {
    return this.read<boolean>('verify', [v, proof])
  }

  /** The EIP-712 digest an executor signs to authorize a relayer. */
  async verdictDigest(v: Verdict): Promise<Hex> {
    return this.read<Hex>('verdictDigest', [v])
  }

  /** Whether the verdict's nullifier has been burned for the domain. */
  async isConsumed(domainId: Hex, nullifier: Hex): Promise<boolean> {
    return this.read<boolean>('isConsumed', [domainId, nullifier])
  }

  /** ERC-165: does the contract advertise IConfidentialPolicyVerdict? */
  async supportsConfidentialPolicyVerdict(): Promise<boolean> {
    return this.read<boolean>('supportsInterface', [CONFIDENTIAL_POLICY_VERDICT_INTERFACE_ID])
  }

  /**
   * Verify and burn a verdict directly. The caller must be the executor.
   * Emits VerdictConsumed and returns the receipt.
   */
  async consume(v: Verdict, proof: Hex): Promise<TransactionReceipt> {
    return this.send('consume', [v, proof])
  }

  /**
   * Verify and burn a verdict via a relayer. executorAuth is a valid EIP-712
   * signature by the executor over verdictDigest(v).
   */
  async consumeRelayed(v: Verdict, proof: Hex, executorAuth: Hex): Promise<TransactionReceipt> {
    return this.send('consume', [v, proof, executorAuth])
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

/**
 * ERC-8354 recommended companion registry. Manages policy domains and roots.
 * Read-only view calls for the guard's checks: domain, currentRoot, and
 * isRootAcceptable.
 */
export class PolicyDomainRegistryClient {
  private readonly publicClient: PublicClient
  private readonly address: Address
  private readonly abi = policyDomainRegistryAbi

  constructor(config: PolicyDomainRegistryClientConfig) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.address = config.address
  }

  /** The Domain record for a domain id. */
  async domain(domainId: Hex): Promise<PolicyDomain> {
    return this.read<PolicyDomain>('domain', [domainId])
  }

  /** The current root, version, and update timestamp for a domain id. */
  async currentRoot(domainId: Hex): Promise<{ root: Hex; version: bigint; updatedAt: bigint }> {
    const result = await this.read<[Hex, bigint, bigint]>('currentRoot', [domainId])
    return { root: result[0], version: result[1], updatedAt: result[2] }
  }

  /** Whether a root is current or superseded within the grace window. */
  async isRootAcceptable(domainId: Hex, root: Hex): Promise<boolean> {
    return this.read<boolean>('isRootAcceptable', [domainId, root])
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
