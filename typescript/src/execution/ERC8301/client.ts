import {
  createPublicClient,
  createWalletClient,
  http,
  parseEventLogs,
  type Account,
  type Address,
  type Hex,
  type PublicClient,
  type TransactionReceipt,
  type WalletClient,
} from 'viem'
import { foundry } from 'viem/chains'
import { agentWorkflowAbi } from './abi.js'
import type { AgentReply, AgentTask, AgentWorkflowClientConfig, RunStatus } from './types.js'

export interface RunResult {
  status: RunStatus
  finalTaskHash: Hex
  completedAt: bigint
}

export interface TaskResult {
  task: AgentTask
  proven: boolean
}

export interface ReplyResult {
  reply: AgentReply
  verifier: Address
  proven: boolean
  verificationDigest: Hex
}

export class AgentWorkflowClient {
  private readonly publicClient: PublicClient
  private readonly walletClient: WalletClient
  private readonly address: Address
  private readonly abi = agentWorkflowAbi

  constructor(config: AgentWorkflowClientConfig, account: Account) {
    const transport = http(config.rpcUrl)
    this.publicClient = createPublicClient({ chain: foundry, transport })
    this.walletClient = createWalletClient({ chain: foundry, transport, account })
    this.address = config.address
  }

  /**
   * Start a new workflow run.
   * @param inputHash - keccak256 of the input
   * @param input - Input plaintext; MAY be empty ("0x")
   * @param expiresAt - Unix timestamp after which the initial task expires
   * @returns The contract-generated workflowRunId, plus emitted NewAgentTask event details
   */
  async run(inputHash: Hex, input: Hex, expiresAt: bigint): Promise<{ workflowRunId: Hex; taskHash: Hex; stage: number }> {
    const receipt = await this.send('run', [inputHash, input, expiresAt])
    const decoded = parseEventLogs({ abi: this.abi, logs: receipt.logs, eventName: 'NewAgentTask' })
    if (decoded.length === 0) {
      throw new Error('run: NewAgentTask event not found in transaction receipt')
    }
    const { workflowRunId, taskHash, stage } = decoded[0].args
    return {
      workflowRunId: workflowRunId as Hex,
      taskHash: taskHash as Hex,
      stage: stage as number,
    }
  }

  /**
   * Query the result of a run.
   * @param workflowRunId - The run identifier
   * @returns The run status, final task hash, and completion timestamp
   */
  async result(workflowRunId: Hex): Promise<RunResult> {
    const [status, finalTaskHash, completedAt] = await this.read<[number, Hex, bigint]>('result', [workflowRunId])
    return { status: status as RunStatus, finalTaskHash, completedAt }
  }

  /**
   * Returns the stored AgentTask and its proven status.
   * @param taskHash - The task hash
   * @returns The task and whether it is proven
   */
  async getTask(taskHash: Hex): Promise<TaskResult> {
    const [task, proven] = await this.read<[AgentTask, boolean]>('getAgentTask', [taskHash])
    return { task, proven }
  }

  /**
   * Returns the stored AgentReply and its verification status.
   * @param replyHash - The reply hash
   * @returns The reply, verifier, proven status, and verification digest
   */
  async getReply(replyHash: Hex): Promise<ReplyResult> {
    const [reply, verifier, proven, verificationDigest] = await this.read<[AgentReply, Address, boolean, Hex]>(
      'getAgentReply',
      [replyHash],
    )
    return { reply, verifier, proven, verificationDigest }
  }

  /**
   * Agent submits a reply to a dispatched task.
   */
  async onAgentReply(reply: AgentReply): Promise<void> {
    await this.send('onAgentReply', [reply])
  }

  /**
   * Submit a cryptographic proof covering one or more anchored replies.
   */
  async onAgentProve(replyHashes: readonly Hex[], proof: Hex): Promise<void> {
    await this.send('onAgentProve', [replyHashes, proof])
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
