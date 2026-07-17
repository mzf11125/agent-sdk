OBSERVATION_COMMITMENT_ABI = [
    {
        "type": "function",
        "name": "record",
        "stateMutability": "nonpayable",
        "inputs": [{"name": "digest", "type": "bytes32"}],
        "outputs": [],
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
        "name": "Recorded",
        "anonymous": False,
        "inputs": [
            {"name": "digest", "type": "bytes32", "indexed": True},
            {"name": "committer", "type": "address", "indexed": True},
        ],
    },
]
