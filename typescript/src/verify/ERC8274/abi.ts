export const proofVerifierAbi = [
  {
    type: 'function',
    name: 'verify',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'inputHash', type: 'bytes32' },
      { name: 'outputHash', type: 'bytes32' },
      { name: 'metadata', type: 'bytes' },
      { name: 'proof', type: 'bytes' },
    ],
    outputs: [{ name: 'valid', type: 'bool' }],
  },
  {
    type: 'function',
    name: 'proofSystem',
    stateMutability: 'view',
    inputs: [],
    outputs: [{ name: '', type: 'string' }],
  },
  {
    type: 'function',
    name: 'proofProfile',
    stateMutability: 'view',
    inputs: [],
    outputs: [{ name: '', type: 'bytes32' }],
  },
] as const

export const agentVerifierAbi = [
  {
    type: 'function',
    name: 'verify',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'taskId', type: 'bytes32' },
      { name: 'agentId', type: 'bytes32' },
      { name: 'inputHash', type: 'bytes32' },
      { name: 'outputHash', type: 'bytes32' },
      { name: 'proof', type: 'bytes' },
    ],
    outputs: [
      { name: 'valid', type: 'bool' },
      { name: 'verificationDigest', type: 'bytes32' },
    ],
  },
  {
    type: 'event',
    name: 'VerificationCompleted',
    inputs: [
      { name: 'taskId', type: 'bytes32', indexed: true },
      { name: 'agentId', type: 'bytes32', indexed: true },
      { name: 'inputHash', type: 'bytes32', indexed: false },
      { name: 'outputHash', type: 'bytes32', indexed: false },
      { name: 'valid', type: 'bool', indexed: false },
      { name: 'verificationDigest', type: 'bytes32', indexed: false },
    ],
  },
] as const

export const agentVerifiableAbi = [
  {
    type: 'function',
    name: 'agentVerifier',
    stateMutability: 'view',
    inputs: [],
    outputs: [{ name: '', type: 'address' }],
  },
  {
    type: 'event',
    name: 'AgentVerifierUpdated',
    inputs: [
      { name: 'oldVerifier', type: 'address', indexed: true },
      { name: 'newVerifier', type: 'address', indexed: true },
    ],
  },
] as const
