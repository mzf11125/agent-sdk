export const verdictTuple = {
  type: 'tuple',
  components: [
    { name: 'agentId', type: 'uint256' },
    { name: 'domainId', type: 'bytes32' },
    { name: 'policyRoot', type: 'bytes32' },
    { name: 'actionCommitment', type: 'bytes32' },
    { name: 'executor', type: 'address' },
    { name: 'expiry', type: 'uint64' },
    { name: 'nullifier', type: 'bytes32' },
    { name: 'decision', type: 'uint8' },
    { name: 'policyKind', type: 'uint8' },
  ],
} as const

export const confidentialPolicyVerdictAbi = [
  {
    type: 'function',
    name: 'verify',
    stateMutability: 'view',
    inputs: [
      { name: 'v', ...verdictTuple },
      { name: 'proof', type: 'bytes' },
    ],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    type: 'function',
    name: 'verdictDigest',
    stateMutability: 'view',
    inputs: [{ name: 'v', ...verdictTuple }],
    outputs: [{ name: '', type: 'bytes32' }],
  },
  {
    type: 'function',
    name: 'consume',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'v', ...verdictTuple },
      { name: 'proof', type: 'bytes' },
    ],
    outputs: [],
  },
  {
    type: 'function',
    name: 'consume',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'v', ...verdictTuple },
      { name: 'proof', type: 'bytes' },
      { name: 'executorAuth', type: 'bytes' },
    ],
    outputs: [],
  },
  {
    type: 'function',
    name: 'isConsumed',
    stateMutability: 'view',
    inputs: [
      { name: 'domainId', type: 'bytes32' },
      { name: 'nullifier', type: 'bytes32' },
    ],
    outputs: [{ name: '', type: 'bool' }],
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
    name: 'VerdictConsumed',
    inputs: [
      { name: 'nullifier', type: 'bytes32', indexed: true },
      { name: 'agentId', type: 'uint256', indexed: true },
      { name: 'domainId', type: 'bytes32', indexed: true },
      { name: 'policyRoot', type: 'bytes32', indexed: false },
      { name: 'actionCommitment', type: 'bytes32', indexed: false },
    ],
  },
] as const

const domainTuple = {
  type: 'tuple',
  components: [
    { name: 'registrar', type: 'address' },
    { name: 'identityRegistry', type: 'address' },
    { name: 'verifier', type: 'address' },
    { name: 'programKey', type: 'bytes32' },
    { name: 'maxRootAge', type: 'uint64' },
    { name: 'active', type: 'bool' },
  ],
} as const

export const policyDomainRegistryAbi = [
  {
    type: 'function',
    name: 'domain',
    stateMutability: 'view',
    inputs: [{ name: 'domainId', type: 'bytes32' }],
    outputs: [{ name: '', ...domainTuple }],
  },
  {
    type: 'function',
    name: 'currentRoot',
    stateMutability: 'view',
    inputs: [{ name: 'domainId', type: 'bytes32' }],
    outputs: [
      { name: 'root', type: 'bytes32' },
      { name: 'version', type: 'uint64' },
      { name: 'updatedAt', type: 'uint64' },
    ],
  },
  {
    type: 'function',
    name: 'isRootAcceptable',
    stateMutability: 'view',
    inputs: [
      { name: 'domainId', type: 'bytes32' },
      { name: 'root', type: 'bytes32' },
    ],
    outputs: [{ name: '', type: 'bool' }],
  },
] as const
