AGENT_SOURCE_BINDING_VIEW_ABI = [
    {
        "type": "function",
        "name": "getSourceNFT",
        "stateMutability": "view",
        "inputs": [{"name": "agentId", "type": "uint256"}],
        "outputs": [
            {"name": "sourceContract", "type": "address"},
            {"name": "sourceTokenId", "type": "uint256"},
        ],
    },
    {
        "type": "function",
        "name": "hasSourceNFT",
        "stateMutability": "view",
        "inputs": [{"name": "agentId", "type": "uint256"}],
        "outputs": [{"name": "", "type": "bool"}],
    },
    {
        "type": "function",
        "name": "isSourceNFTOwnershipValid",
        "stateMutability": "view",
        "inputs": [{"name": "agentId", "type": "uint256"}],
        "outputs": [{"name": "", "type": "bool"}],
    },
    {
        "type": "function",
        "name": "supportsInterface",
        "stateMutability": "view",
        "inputs": [{"name": "interfaceId", "type": "bytes4"}],
        "outputs": [{"name": "", "type": "bool"}],
    },
]
