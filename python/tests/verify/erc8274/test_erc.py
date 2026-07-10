from eth_abi import encode
from eth_account import Account
from web3 import Web3

from agent_sdk.verify.erc8274.client import AgentVerifierClient, ProofVerifierClient, get_trusted_verifier


def _make_clients(deploy_contracts, anvil_rpc_url, anvil_account):
    proof_verifier_address, agent_verifier_address, agent_verifiable_address = deploy_contracts(
        "verify/ERC8274", "DeployERC8274"
    )
    account = Account.from_key(anvil_account(1)["privateKey"])
    proof_verifier_client = ProofVerifierClient(anvil_rpc_url, proof_verifier_address, account)
    agent_verifier_client = AgentVerifierClient(anvil_rpc_url, agent_verifier_address, account)
    return proof_verifier_client, agent_verifier_client, agent_verifier_address, agent_verifiable_address


def test_exposes_proof_system_identifier_and_profile(deploy_contracts, anvil_rpc_url, anvil_account):
    proof_verifier_client, _, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)

    assert proof_verifier_client.proof_system() == "mock-test-only"
    assert proof_verifier_client.proof_profile() == Web3.keccak(text="mock-test-only-v1")


def test_accepts_valid_proof_and_rejects_invalid_one(deploy_contracts, anvil_rpc_url, anvil_account):
    proof_verifier_client, _, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)

    input_hash = Web3.keccak(text="input")
    output_hash = Web3.keccak(text="output")
    metadata = b""
    valid_proof = Web3.solidity_keccak(["bytes32", "bytes32", "bytes"], [input_hash, output_hash, metadata])
    invalid_proof = Web3.keccak(text="garbage")

    assert proof_verifier_client.verify(input_hash, output_hash, metadata, valid_proof) is True
    assert proof_verifier_client.verify(input_hash, output_hash, metadata, invalid_proof) is False


def test_routes_through_agent_verifier_and_produces_recomputable_digest(deploy_contracts, anvil_rpc_url, anvil_account):
    proof_verifier_client, agent_verifier_client, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)

    task_id = Web3.keccak(text="task-1")
    agent_id = Web3.keccak(text="agent-1")
    input_hash = Web3.keccak(text="input")
    output_hash = Web3.keccak(text="output")
    # MockAgentVerifier always routes to the proof verifier with empty
    # metadata (the ERC doesn't specify per-agent metadata), so the proof
    # must be constructed against empty metadata to be accepted here.
    proof = Web3.solidity_keccak(["bytes32", "bytes32", "bytes"], [input_hash, output_hash, b""])

    valid, verification_digest = agent_verifier_client.verify(task_id, agent_id, input_hash, output_hash, proof)

    assert valid is True

    # The ERC's digest formula includes `agentProofProfile`, which isn't
    # exposed generically by IAgentVerifier (see this ERC's README) — but
    # THIS specific mock happens to reuse the proof verifier's own
    # profile, which IS part of the official IProofVerifier API, so we
    # can recompute the expected digest using only supported client calls.
    agent_proof_profile = proof_verifier_client.proof_profile()
    expected_digest = Web3.keccak(
        encode(
            ["bytes32", "bytes32", "bytes32", "bytes32", "bool", "bytes32"],
            [task_id, agent_id, input_hash, output_hash, valid, agent_proof_profile],
        )
    )
    assert verification_digest == expected_digest


def test_returns_invalid_false_for_bad_proof_without_reverting(deploy_contracts, anvil_rpc_url, anvil_account):
    _, agent_verifier_client, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)

    task_id = Web3.keccak(text="task-2")
    agent_id = Web3.keccak(text="agent-2")
    input_hash = Web3.keccak(text="input")
    output_hash = Web3.keccak(text="output")
    bad_proof = Web3.keccak(text="garbage")

    valid, _ = agent_verifier_client.verify(task_id, agent_id, input_hash, output_hash, bad_proof)

    assert valid is False


def test_reads_trusted_verifier_declared_by_settlement_contract(deploy_contracts, anvil_rpc_url, anvil_account):
    _, _, agent_verifier_address, agent_verifiable_address = _make_clients(
        deploy_contracts, anvil_rpc_url, anvil_account
    )

    trusted_verifier = get_trusted_verifier(anvil_rpc_url, agent_verifiable_address)

    assert trusted_verifier.lower() == agent_verifier_address.lower()
