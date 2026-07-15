AGENT_REPUTATION_ABI = [
    {
        "type": "function",
        "name": "getReputation",
        "stateMutability": "view",
        "inputs": [{"name": "agentId", "type": "bytes32"}],
        "outputs": [
            {"name": "completedOrders", "type": "uint64"},
            {"name": "disputedOrders", "type": "uint64"},
            {"name": "totalVolume", "type": "uint64"},
            {"name": "lastActiveAt", "type": "uint64"},
            {"name": "score", "type": "uint16"},
        ],
    },
    {
        "type": "function",
        "name": "getDecayWeight",
        "stateMutability": "view",
        "inputs": [{"name": "agentId", "type": "bytes32"}],
        "outputs": [{"name": "weight", "type": "uint16"}],
    },
    {
        "type": "function",
        "name": "verifyOutcome",
        "stateMutability": "view",
        "inputs": [
            {"name": "orderId", "type": "bytes32"},
            {"name": "proof", "type": "bytes"},
        ],
        "outputs": [{"name": "valid", "type": "bool"}],
    },
]
