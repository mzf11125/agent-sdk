AGENT_WORKFLOW_ABI = [
    {
        "type": "function",
        "name": "run",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "inputHash", "type": "bytes32"},
            {"name": "input", "type": "bytes"},
            {"name": "expiresAt", "type": "uint256"},
        ],
        "outputs": [{"name": "workflowRunId", "type": "bytes32"}],
    },
    {
        "type": "function",
        "name": "result",
        "stateMutability": "view",
        "inputs": [{"name": "workflowRunId", "type": "bytes32"}],
        "outputs": [
            {"name": "status", "type": "uint8"},
            {"name": "finalTaskHash", "type": "bytes32"},
            {"name": "completedAt", "type": "uint256"},
        ],
    },
    {
        "type": "function",
        "name": "getAgentTask",
        "stateMutability": "view",
        "inputs": [{"name": "taskHash", "type": "bytes32"}],
        "outputs": [
            {
                "name": "task",
                "type": "tuple",
                "components": [
                    {"name": "stage", "type": "uint8"},
                    {"name": "taskSeq", "type": "uint256"},
                    {"name": "inputHash", "type": "bytes32"},
                    {"name": "input", "type": "bytes"},
                    {"name": "timestamp", "type": "uint256"},
                    {"name": "expiresAt", "type": "uint256"},
                    {"name": "prevReplyHashes", "type": "bytes32[]"},
                    {"name": "workflowRunId", "type": "bytes32"},
                ],
            },
            {"name": "proven", "type": "bool"},
        ],
    },
    {
        "type": "function",
        "name": "getAgentReply",
        "stateMutability": "view",
        "inputs": [{"name": "replyHash", "type": "bytes32"}],
        "outputs": [
            {
                "name": "reply",
                "type": "tuple",
                "components": [
                    {"name": "outputHash", "type": "bytes32"},
                    {"name": "output", "type": "bytes"},
                    {"name": "timestamp", "type": "uint256"},
                    {"name": "replier", "type": "address"},
                    {"name": "prevTaskHashes", "type": "bytes32[]"},
                    {"name": "workflowRunId", "type": "bytes32"},
                ],
            },
            {"name": "verifier", "type": "address"},
            {"name": "proven", "type": "bool"},
            {"name": "verificationDigest", "type": "bytes32"},
        ],
    },
    {
        "type": "function",
        "name": "onAgentReply",
        "stateMutability": "nonpayable",
        "inputs": [
            {
                "name": "reply",
                "type": "tuple",
                "components": [
                    {"name": "outputHash", "type": "bytes32"},
                    {"name": "output", "type": "bytes"},
                    {"name": "timestamp", "type": "uint256"},
                    {"name": "replier", "type": "address"},
                    {"name": "prevTaskHashes", "type": "bytes32[]"},
                    {"name": "workflowRunId", "type": "bytes32"},
                ],
            },
        ],
        "outputs": [],
    },
    {
        "type": "function",
        "name": "onAgentProve",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "replyHashes", "type": "bytes32[]"},
            {"name": "proof", "type": "bytes"},
        ],
        "outputs": [],
    },
    {
        "type": "event",
        "name": "NewAgentTask",
        "anonymous": False,
        "inputs": [
            {"name": "workflowRunId", "type": "bytes32", "indexed": True},
            {"name": "stage", "type": "uint8", "indexed": True},
            {"name": "taskHash", "type": "bytes32", "indexed": True},
        ],
    },
    {
        "type": "event",
        "name": "WorkflowCompleted",
        "anonymous": False,
        "inputs": [
            {"name": "workflowRunId", "type": "bytes32", "indexed": True},
            {"name": "status", "type": "uint8", "indexed": False},
            {"name": "finalTaskHash", "type": "bytes32", "indexed": False},
            {"name": "timestamp", "type": "uint256", "indexed": False},
        ],
    },
]
