export enum RunStatus {
  Pending = 0,
  Success = 1,
  Failed = 2,
}

export interface AgentTask {
  stage: number
  taskSeq: bigint
  inputHash: `0x${string}`
  input: `0x${string}`
  timestamp: bigint
  expiresAt: bigint
  prevReplyHashes: readonly `0x${string}`[]
  workflowRunId: `0x${string}`
}

export interface AgentReply {
  outputHash: `0x${string}`
  output: `0x${string}`
  timestamp: bigint
  replier: `0x${string}`
  prevTaskHashes: readonly `0x${string}`[]
  workflowRunId: `0x${string}`
}

export interface AgentWorkflowClientConfig {
  rpcUrl: string
  address: `0x${string}`
}
