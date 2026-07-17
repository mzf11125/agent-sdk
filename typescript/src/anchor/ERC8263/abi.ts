export const onChainProofAbi = [
  {
    type: 'function',
    name: 'anchor',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'agentIdScheme', type: 'uint8' },
      { name: 'agentId', type: 'bytes32' },
      { name: 'proofHash', type: 'bytes32' },
    ],
    outputs: [],
  },
  {
    type: 'function',
    name: 'anchorWithAux',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'agentIdScheme', type: 'uint8' },
      { name: 'agentId', type: 'bytes32' },
      { name: 'proofHash', type: 'bytes32' },
      { name: 'aux', type: 'bytes' },
    ],
    outputs: [],
  },
  {
    type: 'event',
    name: 'AnchorProof',
    inputs: [
      { name: 'agentIdScheme', type: 'uint8', indexed: false },
      { name: 'agentId', type: 'bytes32', indexed: true },
      { name: 'proofHash', type: 'bytes32', indexed: true },
      { name: 'operator', type: 'address', indexed: true },
      { name: 'aux', type: 'bytes', indexed: false },
    ],
  },
] as const
