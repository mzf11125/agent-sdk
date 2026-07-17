PROOF_VERIFIER_ABI = [
    {
        "type": "function",
        "name": "verify",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "inputHash", "type": "bytes32"},
            {"name": "outputHash", "type": "bytes32"},
            {"name": "metadata", "type": "bytes"},
            {"name": "proof", "type": "bytes"},
        ],
        "outputs": [{"name": "valid", "type": "bool"}],
    },
    {
        "type": "function",
        "name": "proofSystem",
        "stateMutability": "view",
        "inputs": [],
        "outputs": [{"name": "", "type": "string"}],
    },
    {
        "type": "function",
        "name": "proofProfile",
        "stateMutability": "view",
        "inputs": [],
        "outputs": [{"name": "", "type": "bytes32"}],
    },
]

AGENT_VERIFIER_ABI = [
    {
        "type": "function",
        "name": "verify",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "taskId", "type": "bytes32"},
            {"name": "agentId", "type": "bytes32"},
            {"name": "inputHash", "type": "bytes32"},
            {"name": "outputHash", "type": "bytes32"},
            {"name": "proof", "type": "bytes"},
        ],
        "outputs": [
            {"name": "valid", "type": "bool"},
            {"name": "verificationDigest", "type": "bytes32"},
        ],
    },
    {
        "type": "event",
        "name": "VerificationCompleted",
        "anonymous": False,
        "inputs": [
            {"name": "taskId", "type": "bytes32", "indexed": True},
            {"name": "agentId", "type": "bytes32", "indexed": True},
            {"name": "inputHash", "type": "bytes32", "indexed": False},
            {"name": "outputHash", "type": "bytes32", "indexed": False},
            {"name": "valid", "type": "bool", "indexed": False},
            {"name": "verificationDigest", "type": "bytes32", "indexed": False},
        ],
    },
]

AGENT_VERIFIABLE_ABI = [
    {
        "type": "function",
        "name": "agentVerifier",
        "stateMutability": "view",
        "inputs": [],
        "outputs": [{"name": "", "type": "address"}],
    },
    {
        "type": "event",
        "name": "AgentVerifierUpdated",
        "anonymous": False,
        "inputs": [
            {"name": "oldVerifier", "type": "address", "indexed": True},
            {"name": "newVerifier", "type": "address", "indexed": True},
        ],
    },
]
