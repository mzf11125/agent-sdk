export const boundedAgentActionAbi = [
  {
    type: 'function',
    name: 'registerEnvelope',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'principal', type: 'address' },
      { name: 'capabilityRoot', type: 'bytes32' },
      { name: 'expiresAt', type: 'uint64' },
      { name: 'initData', type: 'bytes' },
    ],
    outputs: [{ name: 'id', type: 'bytes32' }],
  },
  {
    type: 'function',
    name: 'advanceCursor',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'id', type: 'bytes32' },
      { name: 'witness', type: 'bytes' },
    ],
    outputs: [{ name: 'newCursor', type: 'bytes32' }],
  },
  {
    type: 'function',
    name: 'setStatus',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'id', type: 'bytes32' },
      { name: 'newStatus', type: 'uint8' },
    ],
    outputs: [],
  },
  {
    type: 'function',
    name: 'getEnvelope',
    stateMutability: 'view',
    inputs: [{ name: 'id', type: 'bytes32' }],
    outputs: [
      {
        name: 'envelope',
        type: 'tuple',
        components: [
          { name: 'id', type: 'bytes32' },
          { name: 'principal', type: 'address' },
          { name: 'capabilityRoot', type: 'bytes32' },
          { name: 'cursorRoot', type: 'bytes32' },
          { name: 'createdAt', type: 'uint64' },
          { name: 'expiresAt', type: 'uint64' },
          { name: 'status', type: 'uint8' },
        ],
      },
    ],
  },
  {
    type: 'function',
    name: 'getCursor',
    stateMutability: 'view',
    inputs: [{ name: 'id', type: 'bytes32' }],
    outputs: [{ name: '', type: 'bytes32' }],
  },
  {
    type: 'function',
    name: 'getStatus',
    stateMutability: 'view',
    inputs: [{ name: 'id', type: 'bytes32' }],
    outputs: [{ name: '', type: 'uint8' }],
  },
  {
    type: 'function',
    name: 'isActive',
    stateMutability: 'view',
    inputs: [{ name: 'id', type: 'bytes32' }],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    type: 'event',
    name: 'EnvelopeRegistered',
    inputs: [
      { name: 'id', type: 'bytes32', indexed: true },
      { name: 'principal', type: 'address', indexed: true },
      { name: 'capabilityRoot', type: 'bytes32', indexed: true },
    ],
  },
  {
    type: 'event',
    name: 'EnvelopeAdvanced',
    inputs: [
      { name: 'id', type: 'bytes32', indexed: true },
      { name: 'prevCursor', type: 'bytes32', indexed: false },
      { name: 'newCursor', type: 'bytes32', indexed: false },
    ],
  },
  {
    type: 'event',
    name: 'EnvelopeStatusChanged',
    inputs: [
      { name: 'id', type: 'bytes32', indexed: true },
      { name: 'fromStatus', type: 'uint8', indexed: false },
      { name: 'toStatus', type: 'uint8', indexed: false },
    ],
  },
] as const

export const budgetSubstrateAbi = [
  {
    type: 'function',
    name: 'bound',
    stateMutability: 'view',
    inputs: [{ name: 'id', type: 'bytes32' }],
    outputs: [
      { name: 'cap', type: 'uint256' },
      { name: 'asset', type: 'address' },
    ],
  },
  {
    type: 'function',
    name: 'spent',
    stateMutability: 'view',
    inputs: [{ name: 'id', type: 'bytes32' }],
    outputs: [{ name: '', type: 'uint256' }],
  },
  {
    type: 'function',
    name: 'remaining',
    stateMutability: 'view',
    inputs: [{ name: 'id', type: 'bytes32' }],
    outputs: [{ name: '', type: 'uint256' }],
  },
] as const

export const contestableEnvelopeAbi = [
  {
    type: 'function',
    name: 'contest',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'id', type: 'bytes32' },
      { name: 'evidence', type: 'bytes' },
    ],
    outputs: [],
  },
  {
    type: 'function',
    name: 'resolve',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'id', type: 'bytes32' },
      { name: 'outcome', type: 'uint8' },
      { name: 'resolution', type: 'bytes' },
    ],
    outputs: [],
  },
  {
    type: 'event',
    name: 'EnvelopeContested',
    inputs: [
      { name: 'id', type: 'bytes32', indexed: true },
      { name: 'challenger', type: 'address', indexed: true },
    ],
  },
  {
    type: 'event',
    name: 'EnvelopeResolved',
    inputs: [
      { name: 'id', type: 'bytes32', indexed: true },
      { name: 'outcome', type: 'uint8', indexed: false },
    ],
  },
] as const
