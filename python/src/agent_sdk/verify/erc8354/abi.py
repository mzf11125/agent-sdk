VERDICT_COMPONENTS = [
    {"name": "agentId", "type": "uint256"},
    {"name": "domainId", "type": "bytes32"},
    {"name": "policyRoot", "type": "bytes32"},
    {"name": "actionCommitment", "type": "bytes32"},
    {"name": "executor", "type": "address"},
    {"name": "expiry", "type": "uint64"},
    {"name": "nullifier", "type": "bytes32"},
    {"name": "decision", "type": "uint8"},
    {"name": "policyKind", "type": "uint8"},
]

CONFIDENTIAL_POLICY_VERDICT_ABI = [
    {
        "type": "function",
        "name": "verify",
        "stateMutability": "view",
        "inputs": [
            {"name": "v", "type": "tuple", "components": VERDICT_COMPONENTS},
            {"name": "proof", "type": "bytes"},
        ],
        "outputs": [{"name": "", "type": "bool"}],
    },
    {
        "type": "function",
        "name": "verdictDigest",
        "stateMutability": "view",
        "inputs": [
            {"name": "v", "type": "tuple", "components": VERDICT_COMPONENTS},
        ],
        "outputs": [{"name": "", "type": "bytes32"}],
    },
    {
        "type": "function",
        "name": "consume",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "v", "type": "tuple", "components": VERDICT_COMPONENTS},
            {"name": "proof", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "consume",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "v", "type": "tuple", "components": VERDICT_COMPONENTS},
            {"name": "proof", "type": "bytes"},
            {"name": "executorAuth", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "isConsumed",
        "stateMutability": "view",
        "inputs": [
            {"name": "domainId", "type": "bytes32"},
            {"name": "nullifier", "type": "bytes32"},
        ],
        "outputs": [{"name": "", "type": "bool"}],
    },
    {
        "type": "function",
        "name": "supportsInterface",
        "stateMutability": "view",
        "inputs": [{"name": "interfaceId", "type": "bytes4"}],
        "outputs": [{"name": "", "type": "bool"}],
    },
    {
        "type": "event",
        "name": "VerdictConsumed",
        "anonymous": False,
        "inputs": [
            {"name": "nullifier", "type": "bytes32", "indexed": True},
            {"name": "agentId", "type": "uint256", "indexed": True},
            {"name": "domainId", "type": "bytes32", "indexed": True},
            {"name": "policyRoot", "type": "bytes32", "indexed": False},
            {"name": "actionCommitment", "type": "bytes32", "indexed": False},
        ],
    },
]

DOMAIN_COMPONENTS = [
    {"name": "registrar", "type": "address"},
    {"name": "verifier", "type": "address"},
    {"name": "programKey", "type": "bytes32"},
    {"name": "maxRootAge", "type": "uint64"},
    {"name": "active", "type": "bool"},
]

POLICY_DOMAIN_REGISTRY_ABI = [
    {
        "type": "function",
        "name": "domain",
        "stateMutability": "view",
        "inputs": [{"name": "domainId", "type": "bytes32"}],
        "outputs": [{"name": "", "type": "tuple", "components": DOMAIN_COMPONENTS}],
    },
    {
        "type": "function",
        "name": "currentRoot",
        "stateMutability": "view",
        "inputs": [{"name": "domainId", "type": "bytes32"}],
        "outputs": [
            {"name": "root", "type": "bytes32"},
            {"name": "version", "type": "uint64"},
            {"name": "updatedAt", "type": "uint64"},
        ],
    },
    {
        "type": "function",
        "name": "isRootAcceptable",
        "stateMutability": "view",
        "inputs": [
            {"name": "domainId", "type": "bytes32"},
            {"name": "root", "type": "bytes32"},
        ],
        "outputs": [{"name": "", "type": "bool"}],
    },
]
