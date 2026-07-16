export const observationCommitmentAbi = [
  {
    type: 'function',
    name: 'record',
    stateMutability: 'nonpayable',
    inputs: [{ name: 'digest', type: 'bytes32' }],
    outputs: [],
  },
  {
    type: 'function',
    name: 'supportsInterface',
    stateMutability: 'view',
    inputs: [{ name: 'interfaceId', type: 'bytes4' }],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    type: 'event',
    name: 'Recorded',
    inputs: [
      { name: 'digest', type: 'bytes32', indexed: true },
      { name: 'committer', type: 'address', indexed: true },
    ],
  },
] as const
