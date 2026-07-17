ON_CHAIN_PROOF_ABI = [
    {
        "type": "function",
        "name": "anchor",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "agentIdScheme", "type": "uint8"},
            {"name": "agentId", "type": "bytes32"},
            {"name": "proofHash", "type": "bytes32"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "anchorWithAux",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "agentIdScheme", "type": "uint8"},
            {"name": "agentId", "type": "bytes32"},
            {"name": "proofHash", "type": "bytes32"},
            {"name": "aux", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "event",
        "name": "AnchorProof",
        "inputs": [
            {"name": "agentIdScheme", "type": "uint8", "indexed": False},
            {"name": "agentId", "type": "bytes32", "indexed": True},
            {"name": "proofHash", "type": "bytes32", "indexed": True},
            {"name": "operator", "type": "address", "indexed": True},
            {"name": "aux", "type": "bytes", "indexed": False},
        ],
    },
]
