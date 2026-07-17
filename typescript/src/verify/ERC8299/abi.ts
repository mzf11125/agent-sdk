export const wyriweAttestationAbi = [
  {
    type: 'function',
    name: 'verify',
    stateMutability: 'view',
    inputs: [
      {
        name: 'attestation',
        type: 'tuple',
        components: [
          { name: 'agentId', type: 'bytes32' },
          { name: 'registry', type: 'address' },
          { name: 'modelHash', type: 'bytes32' },
          { name: 'rawInputHash', type: 'bytes32' },
          { name: 'sanitizationPipelineHash', type: 'bytes32' },
          { name: 'inputHash', type: 'bytes32' },
          { name: 'outputHash', type: 'bytes32' },
          { name: 'timestamp', type: 'uint256' },
        ],
      },
      { name: 'signature', type: 'bytes' },
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
] as const

export const judgmentExecutionAbi = [
  {
    type: 'function',
    name: 'verify',
    stateMutability: 'view',
    inputs: [
      {
        name: 'attestation',
        type: 'tuple',
        components: [
          { name: 'agentId', type: 'bytes32' },
          { name: 'registry', type: 'address' },
          { name: 'validatorId', type: 'bytes32' },
          { name: 'rawProposalHash', type: 'bytes32' },
          { name: 'verdictHash', type: 'bytes32' },
          { name: 'executedActionHash', type: 'bytes32' },
          { name: 'verdictTimestamp', type: 'uint256' },
          { name: 'executedTimestamp', type: 'uint256' },
          { name: 'recordPointer', type: 'string' },
        ],
      },
      { name: 'signature', type: 'bytes' },
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
] as const
