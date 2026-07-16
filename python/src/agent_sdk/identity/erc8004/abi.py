IDENTITY_REGISTRY_ABI = [
    {
        "type": "function",
        "name": "register",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "agentURI", "type": "string"},
            {
                "name": "metadata",
                "type": "tuple[]",
                "components": [
                    {"name": "metadataKey", "type": "string"},
                    {"name": "metadataValue", "type": "bytes"},
                ],
            },
        ],
        "outputs": [{"name": "agentId", "type": "uint256"}],
    },
    {
        "type": "function",
        "name": "setAgentURI",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "agentId", "type": "uint256"},
            {"name": "agentURI", "type": "string"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "tokenURI",
        "stateMutability": "view",
        "inputs": [{"name": "agentId", "type": "uint256"}],
        "outputs": [{"name": "", "type": "string"}],
    },
    {
        "type": "function",
        "name": "getMetadata",
        "stateMutability": "view",
        "inputs": [
            {"name": "agentId", "type": "uint256"},
            {"name": "metadataKey", "type": "string"},
        ],
        "outputs": [{"name": "", "type": "bytes"}],
    },
    {
        "type": "function",
        "name": "setMetadata",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "agentId", "type": "uint256"},
            {"name": "metadataKey", "type": "string"},
            {"name": "metadataValue", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "setAgentWallet",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "agentId", "type": "uint256"},
            {"name": "newWallet", "type": "address"},
            {"name": "deadline", "type": "uint256"},
            {"name": "signature", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "getAgentWallet",
        "stateMutability": "view",
        "inputs": [{"name": "agentId", "type": "uint256"}],
        "outputs": [{"name": "", "type": "address"}],
    },
    {
        "type": "function",
        "name": "unsetAgentWallet",
        "stateMutability": "nonpayable",
        "inputs": [{"name": "agentId", "type": "uint256"}],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "ownerOf",
        "stateMutability": "view",
        "inputs": [{"name": "tokenId", "type": "uint256"}],
        "outputs": [{"name": "", "type": "address"}],
    },
    {
        "type": "event",
        "name": "Registered",
        "anonymous": False,
        "inputs": [
            {"name": "agentId", "type": "uint256", "indexed": True},
            {"name": "agentURI", "type": "string", "indexed": False},
            {"name": "owner", "type": "address", "indexed": True},
        ],
    },
    # ERC-8323 (Source-Token Agent Binding) — the single spec-defined overload
    # of registerWithSource, for registries deployed bound to a fixed source
    # ERC-721 collection instead of (or in addition to) the base ERC-8004
    # register(). See identity/ERC8323/IAgentSourceBinding.sol in agent-ercs.
    {
        "type": "function",
        "name": "registerWithSource",
        "stateMutability": "payable",
        "inputs": [{"name": "sourceTokenId", "type": "uint256"}],
        "outputs": [{"name": "agentId", "type": "uint256"}],
    },
    # The next two are NOT part of the ERC-8323 base interface -- deployment-
    # specific convenience overloads confirmed live on Merlini's
    # AgentIdentityRegistry (mainnet 0xe0454dfa17a57a84c3e0e2dbfda5318cbbe91e2c,
    # 2026-07-16 Telegram, verified signatures not guessed). Only call these
    # against a registry known to implement this exact extended surface.
    {
        "type": "function",
        "name": "registerWithSource",
        "stateMutability": "payable",
        "inputs": [
            {"name": "agentURI", "type": "string"},
            {"name": "sourceTokenId", "type": "uint256"},
        ],
        "outputs": [{"name": "agentId", "type": "uint256"}],
    },
    {
        "type": "function",
        "name": "registerWithSource",
        "stateMutability": "payable",
        "inputs": [
            {"name": "agentURI", "type": "string"},
            {"name": "sourceTokenId", "type": "uint256"},
            {
                "name": "metadata",
                "type": "tuple[]",
                "components": [
                    {"name": "metadataKey", "type": "string"},
                    {"name": "metadataValue", "type": "bytes"},
                ],
            },
        ],
        "outputs": [{"name": "agentId", "type": "uint256"}],
    },
    {
        "type": "event",
        "name": "SourceNFTLinked",
        "anonymous": False,
        "inputs": [
            {"name": "agentId", "type": "uint256", "indexed": True},
            {"name": "sourceContract", "type": "address", "indexed": True},
            {"name": "sourceTokenId", "type": "uint256", "indexed": False},
        ],
    },
]
