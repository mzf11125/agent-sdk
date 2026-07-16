export const agentSourceBindingAbi = [
  {
    type: 'function',
    name: 'boundCollection',
    stateMutability: 'view',
    inputs: [],
    outputs: [{ name: '', type: 'address' }],
  },
  {
    type: 'function',
    name: 'getSourceNFT',
    stateMutability: 'view',
    inputs: [{ name: 'agentId', type: 'uint256' }],
    outputs: [
      { name: 'sourceContract', type: 'address' },
      { name: 'sourceTokenId', type: 'uint256' },
    ],
  },
  {
    type: 'function',
    name: 'hasSourceNFT',
    stateMutability: 'view',
    inputs: [{ name: 'agentId', type: 'uint256' }],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    type: 'function',
    name: 'isSourceNFTOwnershipValid',
    stateMutability: 'view',
    inputs: [{ name: 'agentId', type: 'uint256' }],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    type: 'function',
    name: 'registerWithSource',
    stateMutability: 'payable',
    inputs: [{ name: 'sourceTokenId', type: 'uint256' }],
    outputs: [{ name: 'agentId', type: 'uint256' }],
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
    name: 'SourceNFTLinked',
    inputs: [
      { name: 'agentId', type: 'uint256', indexed: true },
      { name: 'sourceContract', type: 'address', indexed: true },
      { name: 'sourceTokenId', type: 'uint256', indexed: false },
    ],
  },
] as const
