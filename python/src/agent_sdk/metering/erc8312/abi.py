"""ABI constants for ERC-8312 interfaces."""

BOUNDED_AGENT_ACTION_ABI = [
    {
        "type": "function",
        "name": "registerEnvelope",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "principal", "type": "address"},
            {"name": "capabilityRoot", "type": "bytes32"},
            {"name": "expiresAt", "type": "uint64"},
            {"name": "initData", "type": "bytes"},
        ],
        "outputs": [{"name": "id", "type": "bytes32"}],
    },
    {
        "type": "function",
        "name": "advanceCursor",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "id", "type": "bytes32"},
            {"name": "witness", "type": "bytes"},
        ],
        "outputs": [{"name": "newCursor", "type": "bytes32"}],
    },
    {
        "type": "function",
        "name": "setStatus",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "id", "type": "bytes32"},
            {"name": "newStatus", "type": "uint8"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "getEnvelope",
        "stateMutability": "view",
        "inputs": [{"name": "id", "type": "bytes32"}],
        "outputs": [
            {
                "name": "envelope",
                "type": "tuple",
                "components": [
                    {"name": "id", "type": "bytes32"},
                    {"name": "principal", "type": "address"},
                    {"name": "capabilityRoot", "type": "bytes32"},
                    {"name": "cursorRoot", "type": "bytes32"},
                    {"name": "createdAt", "type": "uint64"},
                    {"name": "expiresAt", "type": "uint64"},
                    {"name": "status", "type": "uint8"},
                ],
            }
        ],
    },
    {
        "type": "function",
        "name": "getCursor",
        "stateMutability": "view",
        "inputs": [{"name": "id", "type": "bytes32"}],
        "outputs": [{"name": "", "type": "bytes32"}],
    },
    {
        "type": "function",
        "name": "getStatus",
        "stateMutability": "view",
        "inputs": [{"name": "id", "type": "bytes32"}],
        "outputs": [{"name": "", "type": "uint8"}],
    },
    {
        "type": "function",
        "name": "isActive",
        "stateMutability": "view",
        "inputs": [{"name": "id", "type": "bytes32"}],
        "outputs": [{"name": "", "type": "bool"}],
    },
    {
        "type": "event",
        "name": "EnvelopeRegistered",
        "anonymous": False,
        "inputs": [
            {"name": "id", "type": "bytes32", "indexed": True},
            {"name": "principal", "type": "address", "indexed": True},
            {"name": "capabilityRoot", "type": "bytes32", "indexed": True},
        ],
    },
    {
        "type": "event",
        "name": "EnvelopeAdvanced",
        "anonymous": False,
        "inputs": [
            {"name": "id", "type": "bytes32", "indexed": True},
            {"name": "prevCursor", "type": "bytes32", "indexed": False},
            {"name": "newCursor", "type": "bytes32", "indexed": False},
        ],
    },
    {
        "type": "event",
        "name": "EnvelopeStatusChanged",
        "anonymous": False,
        "inputs": [
            {"name": "id", "type": "bytes32", "indexed": True},
            {"name": "fromStatus", "type": "uint8", "indexed": False},
            {"name": "toStatus", "type": "uint8", "indexed": False},
        ],
    },
]

BUDGET_SUBSTRATE_ABI = [
    {
        "type": "function",
        "name": "bound",
        "stateMutability": "view",
        "inputs": [{"name": "id", "type": "bytes32"}],
        "outputs": [
            {"name": "cap", "type": "uint256"},
            {"name": "asset", "type": "address"},
        ],
    },
    {
        "type": "function",
        "name": "spent",
        "stateMutability": "view",
        "inputs": [{"name": "id", "type": "bytes32"}],
        "outputs": [{"name": "", "type": "uint256"}],
    },
    {
        "type": "function",
        "name": "remaining",
        "stateMutability": "view",
        "inputs": [{"name": "id", "type": "bytes32"}],
        "outputs": [{"name": "", "type": "uint256"}],
    },
]

CONTESTABLE_ENVELOPE_ABI = [
    {
        "type": "function",
        "name": "contest",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "id", "type": "bytes32"},
            {"name": "evidence", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "resolve",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "id", "type": "bytes32"},
            {"name": "outcome", "type": "uint8"},
            {"name": "resolution", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "event",
        "name": "EnvelopeContested",
        "anonymous": False,
        "inputs": [
            {"name": "id", "type": "bytes32", "indexed": True},
            {"name": "challenger", "type": "address", "indexed": True},
        ],
    },
    {
        "type": "event",
        "name": "EnvelopeResolved",
        "anonymous": False,
        "inputs": [
            {"name": "id", "type": "bytes32", "indexed": True},
            {"name": "outcome", "type": "uint8", "indexed": False},
        ],
    },
]
